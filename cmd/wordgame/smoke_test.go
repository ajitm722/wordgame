package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/fleetdm/wordgame/internal/handler"
	"github.com/fleetdm/wordgame/internal/store"
)

// setupSmoke starts a real HTTP server with a deterministic word list.
// The word "ZZZZ" means 'Z' is always correct and any other letter is wrong.
// Returns the server URL. The server is automatically cleaned up at end of test.
func setupSmoke(t *testing.T) string {
	t.Helper()

	store := store.NewGameStore()
	srv := handler.NewServer(store, []string{"ZZZZ"})

	r := mux.NewRouter()
	r.HandleFunc("/new", srv.HandleNewGame).Methods(http.MethodPost)
	r.HandleFunc("/guess", srv.HandleGuess).Methods(http.MethodPost)

	ts := httptest.NewServer(r)
	t.Cleanup(ts.Close)
	return ts.URL
}

// postJSON is a helper that sends a POST request with an optional JSON body.
func postJSON(t *testing.T, url, path string, body any) (*http.Response, []byte) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}

	resp, err := http.Post(url+path, "application/json", reqBody)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return resp, respBody
}

// TestSmokeNewGame_Shape verifies that POST /new returns the correct shape.
func TestSmokeNewGame_Shape(t *testing.T) {
	url := setupSmoke(t)

	resp, body := postJSON(t, url, "/new", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /new status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, body)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var game handler.NewGameResponse
	if err := json.Unmarshal(body, &game); err != nil {
		t.Fatalf("unmarshal response: %v\nbody: %s", err, body)
	}

	if game.ID == "" {
		t.Error("game ID is empty")
	}
	if game.Current == "" {
		t.Error("current board is empty")
	}
	if game.GuessesRemaining != 6 {
		t.Errorf("guesses_remaining = %d, want 6", game.GuessesRemaining)
	}
}

// TestSmokeGuess_Correct verifies a correct guess updates the board.
func TestSmokeGuess_Correct(t *testing.T) {
	url := setupSmoke(t)

	// Create a new game
	_, newBody := postJSON(t, url, "/new", nil)
	var newGame handler.NewGameResponse
	if err := json.Unmarshal(newBody, &newGame); err != nil {
		t.Fatalf("unmarshal new game: %v", err)
	}

	// Guess 'Z' — this is in "ZZZZ", so it's correct
	resp, body := postJSON(t, url, "/guess", handler.GuessRequest{
		ID:    newGame.ID,
		Guess: "Z",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /guess status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, body)
	}

	var guessResp handler.GuessResponse
	if err := json.Unmarshal(body, &guessResp); err != nil {
		t.Fatalf("unmarshal guess response: %v\nbody: %s", err, body)
	}

	// Board should have changed (all underscores → mostly Z's)
	if guessResp.Current == newGame.Current {
		t.Errorf("board unchanged after correct guess: %q", guessResp.Current)
	}
	// Guesses should NOT have decreased for a correct guess
	if guessResp.GuessesRemaining != newGame.GuessesRemaining {
		t.Errorf("guesses_remaining changed from %d to %d after correct guess",
			newGame.GuessesRemaining, guessResp.GuessesRemaining)
	}
}

// TestSmokeGuess_Wrong verifies a wrong guess decrements guesses.
func TestSmokeGuess_Wrong(t *testing.T) {
	url := setupSmoke(t)

	// Create a new game
	_, newBody := postJSON(t, url, "/new", nil)
	var newGame handler.NewGameResponse
	if err := json.Unmarshal(newBody, &newGame); err != nil {
		t.Fatalf("unmarshal new game: %v", err)
	}

	// Guess 'A' — this is NOT in "ZZZZ", so it's wrong
	resp, body := postJSON(t, url, "/guess", handler.GuessRequest{
		ID:    newGame.ID,
		Guess: "A",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /guess status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, body)
	}

	var guessResp handler.GuessResponse
	if err := json.Unmarshal(body, &guessResp); err != nil {
		t.Fatalf("unmarshal guess response: %v\nbody: %s", err, body)
	}

	// Board should be unchanged
	if guessResp.Current != newGame.Current {
		t.Errorf("board changed after wrong guess: %q → %q", newGame.Current, guessResp.Current)
	}
	// Guesses should have decreased by 1
	if guessResp.GuessesRemaining != newGame.GuessesRemaining-1 {
		t.Errorf("guesses_remaining = %d, want %d",
			guessResp.GuessesRemaining, newGame.GuessesRemaining-1)
	}
}

// TestSmokeGuess_DeletedGame verifies that guessing on a completed game returns 404.
func TestSmokeGuess_DeletedGame(t *testing.T) {
	url := setupSmoke(t)

	// Create a new game
	_, newBody := postJSON(t, url, "/new", nil)
	var newGame handler.NewGameResponse
	if err := json.Unmarshal(newBody, &newGame); err != nil {
		t.Fatalf("unmarshal new game: %v", err)
	}

	// Make 6 wrong guesses to exhaust the game (word is "ZZZZ", guess 'A')
	for i := 0; i < 6; i++ {
		resp, body := postJSON(t, url, "/guess", handler.GuessRequest{
			ID:    newGame.ID,
			Guess: "A",
		})

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("guess %d: status = %d, want 200\nbody: %s", i+1, resp.StatusCode, body)
		}

		var guessResp handler.GuessResponse
		if err := json.Unmarshal(body, &guessResp); err != nil {
			t.Fatalf("guess %d: unmarshal response: %v\nbody: %s", i+1, err, body)
		}

		// On the 6th guess, the word should be revealed
		if i == 5 {
			if guessResp.Word != "ZZZZ" {
				t.Errorf("lost game: word = %q, want %q", guessResp.Word, "ZZZZ")
			}
			if guessResp.GuessesRemaining != 0 {
				t.Errorf("lost game: guesses_remaining = %d, want 0", guessResp.GuessesRemaining)
			}
		}
	}

	// The game should now be deleted — any further guess returns 404
	resp, body := postJSON(t, url, "/guess", handler.GuessRequest{
		ID:    newGame.ID,
		Guess: "A",
	})

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("POST /guess after completion: status = %d, want 404\nbody: %s",
			resp.StatusCode, body)
	}

	var errResp handler.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		t.Fatalf("unmarshal error response: %v\nbody: %s", err, body)
	}
	if errResp.Error != "game not found" {
		t.Errorf("error message = %q, want %q", errResp.Error, "game not found")
	}
}
