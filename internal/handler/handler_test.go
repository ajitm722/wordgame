package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fleetdm/wordgame/internal/game"
	"github.com/fleetdm/wordgame/internal/store"
)

func TestHandleNewGame(t *testing.T) {
	s := store.NewGameStore()
	words := []string{"APPLE", "ORANGE", "BANANA"}
	srv := NewServer(s, words)

	req := httptest.NewRequest(http.MethodPost, "/new", nil)
	rec := httptest.NewRecorder()

	srv.HandleNewGame(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp NewGameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID == "" {
		t.Error("id should not be empty")
	}
	if resp.GuessesRemaining != 6 {
		t.Errorf("guesses_remaining = %d, want 6", resp.GuessesRemaining)
	}
	// Current should be all underscores matching one of our words
	if len(resp.Current) == 0 {
		t.Error("current should not be empty")
	}
	for _, ch := range resp.Current {
		if ch != '_' {
			t.Errorf("current should only contain underscores, got %q", resp.Current)
			break
		}
	}
}

func TestHandleNewGame_MethodNotAllowed(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	req := httptest.NewRequest(http.MethodGet, "/new", nil)
	rec := httptest.NewRecorder()

	srv.HandleNewGame(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleGuess_Correct(t *testing.T) {
	s := store.NewGameStore()
	words := []string{"APPLE"}
	srv := NewServer(s, words)

	// First create a game
	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// Make a correct guess
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()

	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Current != "A____" {
		t.Errorf("current = %q, want %q", resp.Current, "A____")
	}
	if resp.GuessesRemaining != 6 {
		t.Errorf("guesses_remaining = %d, want 6", resp.GuessesRemaining)
	}
}

func TestHandleGuess_Wrong(t *testing.T) {
	s := store.NewGameStore()
	words := []string{"APPLE"}
	srv := NewServer(s, words)

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()

	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Current != "_____" {
		t.Errorf("current = %q, want %q", resp.Current, "_____")
	}
	if resp.GuessesRemaining != 5 {
		t.Errorf("guesses_remaining = %d, want 5", resp.GuessesRemaining)
	}
}

func TestHandleGuess_GameNotFound(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	body := strings.NewReader(`{"id":"nonexistent","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec := httptest.NewRecorder()

	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "game not found" {
		t.Errorf("error = %q, want %q", errResp.Error, "game not found")
	}
}

func TestHandleGuess_AlreadyCompleted(t *testing.T) {
	s := store.NewGameStore()
	words := []string{"A"} // single-letter word for instant win
	srv := NewServer(s, words)

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// Win the game — word should be revealed
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	var winResp GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &winResp)
	if winResp.Word != "A" {
		t.Errorf("word should be revealed on win, got %q", winResp.Word)
	}

	// Game should be deleted — further guesses return "game not found"
	body = strings.NewReader(`{"id":"` + newResp.ID + `","guess":"B"}`)
	req = httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "game not found" {
		t.Errorf("error = %q, want %q (game should be deleted after completion)", errResp.Error, "game not found")
	}
}

func TestHandleGuess_InvalidGuess(t *testing.T) {
	s := store.NewGameStore()
	words := []string{"APPLE"}
	srv := NewServer(s, words)

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	tests := []struct {
		name  string
		guess string
		want  string
	}{
		{"empty", "", "missing guess"},
		{"too long", "AB", "guess must be a single character"},
		{"digit", "5", "guess must be a single letter A-Z"},
		{"special", "@", "guess must be a single letter A-Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"` + tt.guess + `"}`)
			req := httptest.NewRequest(http.MethodPost, "/guess", body)
			rec := httptest.NewRecorder()

			srv.HandleGuess(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
			}

			var errResp ErrorResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
			if errResp.Error != tt.want {
				t.Errorf("error = %q, want %q", errResp.Error, tt.want)
			}
		})
	}
}

func TestHandleGuess_MissingID(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	body := strings.NewReader(`{"guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec := httptest.NewRecorder()

	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "missing game id" {
		t.Errorf("error = %q, want %q", errResp.Error, "missing game id")
	}
}

func TestHandleGuess_InvalidJSON(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec := httptest.NewRecorder()

	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "invalid request body" {
		t.Errorf("error = %q, want %q", errResp.Error, "invalid request body")
	}
}

func TestHandleGuess_MethodNotAllowed(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	req := httptest.NewRequest(http.MethodGet, "/guess", nil)
	rec := httptest.NewRecorder()

	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleGuess_DuplicateGuess(t *testing.T) {
	s := store.NewGameStore()
	words := []string{"APPLE"}
	srv := NewServer(s, words)

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// First wrong guess
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	var resp1 GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp1)

	// Duplicate same wrong guess — must decrement again
	body = strings.NewReader(`{"id":"` + newResp.ID + `","guess":"Z"}`)
	req = httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	var resp2 GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp2)

	// Repeat wrong guess: 5 → 4 (every guess counts)
	if resp2.GuessesRemaining != resp1.GuessesRemaining-1 {
		t.Errorf("guesses_remaining = %d, want %d (repeat wrong guess should still decrement)",
			resp2.GuessesRemaining, resp1.GuessesRemaining-1)
	}
}

// --- JSON response helper tests ---

func TestWriteError_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "test error")

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", w.Header().Get("Content-Type"))
	}
}

func TestWriteJSON_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", w.Header().Get("Content-Type"))
	}
}

// --- Postel's Law normalisation tests ---

func TestNormaliseGuess(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already uppercase", "A", "A"},
		{"lowercase", "a", "A"},
		{"mixed case", "AbC", "ABC"},
		{"leading whitespace", "  A", "A"},
		{"trailing whitespace", "A  ", "A"},
		{"surrounding whitespace", "  A  ", "A"},
		{"whitespace only", "   ", ""},
		{"empty", "", ""},
		{"multiple letters", "APPLE", "APPLE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normaliseGuess(tt.input)
			if got != tt.want {
				t.Errorf("normaliseGuess(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateGuess(t *testing.T) {
	tests := []struct {
		name    string
		guess   string
		wantErr string
	}{
		{"valid letter", "A", ""},
		{"valid letter Z", "Z", ""},
		{"valid letter M", "M", ""},
		{"empty string", "", "missing guess"},
		{"too long", "AB", "guess must be a single character"},
		{"digit", "5", "guess must be a single letter A-Z"},
		{"special char", "@", "guess must be a single letter A-Z"},
		{"lowercase after normalisation already done", "a", "guess must be a single letter A-Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGuess(tt.guess)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validateGuess(%q) = %v, want nil", tt.guess, err)
				}
			} else {
				if err == nil {
					t.Errorf("validateGuess(%q) = nil, want %q", tt.guess, tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("validateGuess(%q) = %q, want %q", tt.guess, err.Error(), tt.wantErr)
				}
			}
		})
	}
}

// --- decodeJSONBody tests ---

func TestDecodeJSONBody_Valid(t *testing.T) {
	body := strings.NewReader(`{"id":"abc","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)

	var result GuessRequest
	if err := decodeJSONBody(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc" {
		t.Errorf("ID = %q, want %q", result.ID, "abc")
	}
	if result.Guess != "A" {
		t.Errorf("Guess = %q, want %q", result.Guess, "A")
	}
}

func TestDecodeJSONBody_UnknownFields(t *testing.T) {
	body := strings.NewReader(`{"id":"abc","guess":"A","extra":"bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)

	var result GuessRequest
	if err := decodeJSONBody(req, &result); err == nil {
		t.Error("expected error for unknown fields")
	}
}

func TestDecodeJSONBody_InvalidJSON(t *testing.T) {
	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)

	var result GuessRequest
	if err := decodeJSONBody(req, &result); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodeJSONBody_PartialJSON(t *testing.T) {
	body := strings.NewReader(`{"id":"abc"}`) // guess missing — valid JSON, just partial
	req := httptest.NewRequest(http.MethodPost, "/guess", body)

	var result GuessRequest
	if err := decodeJSONBody(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc" {
		t.Errorf("ID = %q, want %q", result.ID, "abc")
	}
	if result.Guess != "" {
		t.Errorf("Guess = %q, want empty", result.Guess)
	}
}

// --- Integration tests ---

// TestHandleNewGame_MultipleCreatesIndependent verifies that
// creating multiple games produces independent states.
func TestHandleNewGame_MultipleCreatesIndependent(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE", "ORANGE"})

	// Create two games
	rec1 := httptest.NewRecorder()
	srv.HandleNewGame(rec1, httptest.NewRequest(http.MethodPost, "/new", nil))
	var r1 NewGameResponse
	_ = json.Unmarshal(rec1.Body.Bytes(), &r1)

	rec2 := httptest.NewRecorder()
	srv.HandleNewGame(rec2, httptest.NewRequest(http.MethodPost, "/new", nil))
	var r2 NewGameResponse
	_ = json.Unmarshal(rec2.Body.Bytes(), &r2)

	// IDs must be unique
	if r1.ID == r2.ID {
		t.Error("expected unique game IDs")
	}

	// Both start with 6 guesses and all underscores
	if r1.GuessesRemaining != 6 || r2.GuessesRemaining != 6 {
		t.Error("both games should start with 6 guesses")
	}
	for _, ch := range r1.Current {
		if ch != '_' {
			t.Errorf("game 1 current should be underscores, got %q", r1.Current)
			break
		}
	}
	for _, ch := range r2.Current {
		if ch != '_' {
			t.Errorf("game 2 current should be underscores, got %q", r2.Current)
			break
		}
	}
}

// TestHandleGuess_DisallowUnknownFields verifies that requests
// with extra JSON fields are rejected.
func TestHandleGuess_DisallowUnknownFields(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// Send an extra field that doesn't belong
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"A","extra":"bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (unknown fields should be rejected)", rec.Code, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "invalid request body" {
		t.Errorf("error = %q, want %q", errResp.Error, "invalid request body")
	}
}

// TestEndToEnd_FullGameWin simulates a complete game from
// creation to winning by guessing all letters.
// Verifies the word is revealed on win and the game is cleaned up.
func TestEndToEnd_FullGameWin(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"CAT"})

	// 1. Start new game
	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	if newResp.Current != "___" {
		t.Errorf("initial current = %q, want %q", newResp.Current, "___")
	}

	// 2. Guess 'C'
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"C"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	var resp GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Current != "C__" {
		t.Errorf("after 'C': current = %q, want %q", resp.Current, "C__")
	}
	if resp.Word != "" {
		t.Errorf("word should not be revealed mid-game, got %q", resp.Word)
	}

	// 3. Guess 'A'
	body = strings.NewReader(`{"id":"` + newResp.ID + `","guess":"A"}`)
	req = httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Current != "CA_" {
		t.Errorf("after 'A': current = %q, want %q", resp.Current, "CA_")
	}

	// 4. Guess 'T' — should win, reveal word, and delete game
	body = strings.NewReader(`{"id":"` + newResp.ID + `","guess":"T"}`)
	req = httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Current != "CAT" {
		t.Errorf("after 'T': current = %q, want %q", resp.Current, "CAT")
	}
	if resp.Word != "CAT" {
		t.Errorf("word should be revealed on win, got %q", resp.Word)
	}
	if s.Len() != 0 {
		t.Errorf("completed game should be deleted from store, got %d games", s.Len())
	}

	// 5. Game is gone — further guesses return "game not found"
	body = strings.NewReader(`{"id":"` + newResp.ID + `","guess":"X"}`)
	req = httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("deleted game should return 404, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "game not found" {
		t.Errorf("error = %q, want %q", errResp.Error, "game not found")
	}
}

// TestEndToEnd_FullGameLoss simulates a complete game from
// creation to losing by exhausting all guesses.
// Verifies the word is revealed on loss and the game is cleaned up.
func TestEndToEnd_FullGameLoss(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"CAT"})

	// 1. Start new game
	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// 2. Make 6 wrong guesses
	wrongLetters := []string{"Z", "Y", "X", "W", "V", "U"}
	var lastResp GuessResponse
	for i, letter := range wrongLetters {
		body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"` + letter + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/guess", body)
		rec = httptest.NewRecorder()
		srv.HandleGuess(rec, req)
		_ = json.Unmarshal(rec.Body.Bytes(), &lastResp)

		expectedRemaining := 5 - i
		if lastResp.GuessesRemaining != expectedRemaining {
			t.Errorf("after guess %d (%s): guesses_remaining = %d, want %d",
				i+1, letter, lastResp.GuessesRemaining, expectedRemaining)
		}
		// Word should NOT be revealed until the final guess
		if i < 5 && lastResp.Word != "" {
			t.Errorf("word should not be revealed mid-game, got %q", lastResp.Word)
		}
	}

	// 3. Verify loss: word revealed on final guess
	if lastResp.GuessesRemaining != 0 {
		t.Errorf("final guesses_remaining = %d, want 0", lastResp.GuessesRemaining)
	}
	if lastResp.Word != "CAT" {
		t.Errorf("word should be revealed on loss, got %q", lastResp.Word)
	}

	// 4. Game is deleted from store
	if s.Len() != 0 {
		t.Errorf("completed game should be deleted, got %d games", s.Len())
	}

	// 5. Further guesses return "game not found"
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"B"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("deleted game should return 404, got %d", rec.Code)
	}
}

// TestHandleGuess_ConcurrentAccess verifies that concurrent
// guesses on the same game do not cause data races.
func TestHandleGuess_ConcurrentAccess(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"BANANA"})

	// Create a game
	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// Spawn 10 goroutines all guessing 'A' concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"A"}`)
			req := httptest.NewRequest(http.MethodPost, "/guess", body)
			rec := httptest.NewRecorder()
			srv.HandleGuess(rec, req)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify the game state is still consistent —
	// all 'A' positions should be revealed exactly once
	g := s.Get(newResp.ID)
	if g == nil {
		t.Fatal("game should still exist after concurrent guesses")
	}
	if g.Current != "_A_A_A" {
		t.Errorf("expected _A_A_A after concurrent 'A' guesses, got %q", g.Current)
	}
	// GuessesRemaining should still be 6 (correct guess, no penalties)
	if g.GuessesRemaining != 6 {
		t.Errorf("expected 6 guesses remaining, got %d", g.GuessesRemaining)
	}
	// Verify the game is still in progress and consistent
	if g.Status != game.StatusInProgress {
		t.Errorf("expected StatusInProgress, got %d", g.Status)
	}
}

// TestHandleGuess_ConcurrentDifferentGames verifies that
// concurrent guesses on different games are independent.
func TestHandleGuess_ConcurrentDifferentGames(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE", "ORANGE"})

	// Create two games
	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var r1 NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &r1)

	rec = httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var r2 NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &r2)

	// Concurrently guess on different games
	done := make(chan bool, 2)
	go func() {
		body := strings.NewReader(`{"id":"` + r1.ID + `","guess":"Z"}`)
		req := httptest.NewRequest(http.MethodPost, "/guess", body)
		rec := httptest.NewRecorder()
		srv.HandleGuess(rec, req)
		done <- true
	}()
	go func() {
		body := strings.NewReader(`{"id":"` + r2.ID + `","guess":"P"}`)
		req := httptest.NewRequest(http.MethodPost, "/guess", body)
		rec := httptest.NewRecorder()
		srv.HandleGuess(rec, req)
		done <- true
	}()
	<-done
	<-done

	// Game 1 should have 5 guesses (one wrong)
	g1 := s.Get(r1.ID)
	if g1.GuessesRemaining != 5 {
		t.Errorf("game 1: expected 5 guesses remaining, got %d", g1.GuessesRemaining)
	}

	// Game 2 state depends on which word was chosen
	g2 := s.Get(r2.ID)
	if g2.GuessesRemaining < 5 || g2.GuessesRemaining > 6 {
		t.Errorf("game 2: unexpected guesses_remaining %d", g2.GuessesRemaining)
	}
}

// TestHandleNewGame_PostelsLaw_RequestBodyIgnored verifies that
// any body (or no body) is accepted for POST /new.
func TestHandleNewGame_PostelsLaw_RequestBodyIgnored(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	tests := []struct {
		name string
		body string
	}{
		{"no body", ""},
		{"empty JSON", "{}"},
		{"garbage body", "not json"},
		{"extra fields", `{"foo":"bar"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			var req *http.Request
			if bodyReader != nil {
				req = httptest.NewRequest(http.MethodPost, "/new", bodyReader)
			} else {
				req = httptest.NewRequest(http.MethodPost, "/new", nil)
			}
			rec := httptest.NewRecorder()
			srv.HandleNewGame(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d (body should be ignored)", rec.Code, http.StatusOK)
			}
		})
	}
}

// TestHandleGuess_PostelsLaw_MixedCaseAndWhitespace tests combined
// normalisation: lowercase + surrounding whitespace.
func TestHandleGuess_PostelsLaw_MixedCaseAndWhitespace(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	// Lowercase with trailing/leading whitespace
	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"  p  "}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Current != "_PP__" {
		t.Errorf("current = %q, want %q (lowercase+whitespace 'p' should be normalised)", resp.Current, "_PP__")
	}
}

// TestHandleGuess_IDUnchangedInResponse verifies the handler echoes
// back the same game ID in the response.
func TestHandleGuess_IDUnchangedInResponse(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, httptest.NewRequest(http.MethodPost, "/new", nil))
	var newResp NewGameResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &newResp)

	body := strings.NewReader(`{"id":"` + newResp.ID + `","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec = httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	var resp GuessResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.ID != newResp.ID {
		t.Errorf("response ID = %q, want %q", resp.ID, newResp.ID)
	}
}

// TestHandleNewGame_IdentifierError verifies that when ID generation fails,
// the handler returns a 500 Internal Server Error.
func TestHandleNewGame_IdentifierError(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"}, WithIDGenerator(func() (string, error) {
		return "", errors.New("uuid failure")
	}))

	req := httptest.NewRequest(http.MethodPost, "/new", nil)
	rec := httptest.NewRecorder()
	srv.HandleNewGame(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "failed to generate game ID" {
		t.Errorf("error = %q, want %q", errResp.Error, "failed to generate game ID")
	}
}

// TestHandleGuess_ApplyGuessError_GameAlreadyWon verifies that
// a completed game returns 409 Conflict — the game exists but is in
// a conflicting state (already completed by a concurrent request).
func TestHandleGuess_ApplyGuessError_GameAlreadyWon(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	g := game.NewGame("already-won", "APPLE")
	g.Status = game.StatusWon
	s.Save(g)

	body := strings.NewReader(`{"id":"already-won","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec := httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "game already completed" {
		t.Errorf("error = %q, want %q", errResp.Error, "game already completed")
	}
}

// TestHandleGuess_ApplyGuessError_GameAlreadyLost verifies the same
// 409 Conflict for a game that has already been lost.
func TestHandleGuess_ApplyGuessError_GameAlreadyLost(t *testing.T) {
	s := store.NewGameStore()
	srv := NewServer(s, []string{"APPLE"})

	g := game.NewGame("already-lost", "APPLE")
	g.Status = game.StatusLost
	s.Save(g)

	body := strings.NewReader(`{"id":"already-lost","guess":"A"}`)
	req := httptest.NewRequest(http.MethodPost, "/guess", body)
	rec := httptest.NewRecorder()
	srv.HandleGuess(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error != "game already completed" {
		t.Errorf("error = %q, want %q", errResp.Error, "game already completed")
	}
}
