package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/fleetdm/wordgame/internal/game"
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
	registerRoutes(r, srv)

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

// TestSmokeNewGame_Shape verifies POST /api/v1/new returns the correct JSON shape via real HTTP.
func TestSmokeNewGame_Shape(t *testing.T) {
	url := setupSmoke(t)

	resp, body := postJSON(t, url, "/api/v1/new", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/v1/new status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, body)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var newGame handler.NewGameResponse
	if err := json.Unmarshal(body, &newGame); err != nil {
		t.Fatalf("unmarshal response: %v\nbody: %s", err, body)
	}

	if newGame.ID == "" {
		t.Error("game ID is empty")
	}
	if newGame.Current == "" {
		t.Error("current board is empty")
	}
	if newGame.GuessesRemaining != game.MaxGuesses {
		t.Errorf("guesses_remaining = %d, want %d", newGame.GuessesRemaining, game.MaxGuesses)
	}
}

// TestSmokeGuess_Correct verifies a correct guess ("Z") updates the board without decreasing guesses remaining.
func TestSmokeGuess_Correct(t *testing.T) {
	url := setupSmoke(t)

	// Create a new game
	_, newBody := postJSON(t, url, "/api/v1/new", nil)
	var newGame handler.NewGameResponse
	if err := json.Unmarshal(newBody, &newGame); err != nil {
		t.Fatalf("unmarshal new game: %v", err)
	}

	// Guess 'Z' — this is in "ZZZZ", so it's correct
	resp, body := postJSON(t, url, "/api/v1/guess", handler.GuessRequest{
		ID:    newGame.ID,
		Guess: "Z",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/v1/guess status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, body)
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

// TestSmokeGuess_Wrong verifies a wrong guess ("A") decrements remaining guesses by 1.
func TestSmokeGuess_Wrong(t *testing.T) {
	url := setupSmoke(t)

	// Create a new game
	_, newBody := postJSON(t, url, "/api/v1/new", nil)
	var newGame handler.NewGameResponse
	if err := json.Unmarshal(newBody, &newGame); err != nil {
		t.Fatalf("unmarshal new game: %v", err)
	}

	// Guess 'A' — this is NOT in "ZZZZ", so it's wrong
	resp, body := postJSON(t, url, "/api/v1/guess", handler.GuessRequest{
		ID:    newGame.ID,
		Guess: "A",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/v1/guess status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, body)
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

// TestSmokeGuess_DeletedGame verifies guessing on a completed/exhausted game returns 404 with "game not found" error.
func TestSmokeGuess_DeletedGame(t *testing.T) {
	url := setupSmoke(t)

	// Create a new game
	_, newBody := postJSON(t, url, "/api/v1/new", nil)
	var newGame handler.NewGameResponse
	if err := json.Unmarshal(newBody, &newGame); err != nil {
		t.Fatalf("unmarshal new game: %v", err)
	}

	// Make game.MaxGuesses wrong guesses to exhaust the game (word is "ZZZZ", guess 'A')
	for i := range game.MaxGuesses {
		resp, body := postJSON(t, url, "/api/v1/guess", handler.GuessRequest{
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

		// On the final guess, the word should be revealed
		if i == game.MaxGuesses-1 {
			if guessResp.Word != "ZZZZ" {
				t.Errorf("lost game: word = %q, want %q", guessResp.Word, "ZZZZ")
			}
			if guessResp.GuessesRemaining != 0 {
				t.Errorf("lost game: guesses_remaining = %d, want 0", guessResp.GuessesRemaining)
			}
		}
	}

	// The game should now be deleted — any further guess returns 404
	resp, body := postJSON(t, url, "/api/v1/guess", handler.GuessRequest{
		ID:    newGame.ID,
		Guess: "A",
	})

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("POST /api/v1/guess after completion: status = %d, want 404\nbody: %s",
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
