// Package handler implements HTTP handlers for the word-guessing game API.
// It applies Postel's Law: be liberal in what you accept, conservative in what you send.
package handler

import (
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/fleetdm/wordgame/internal/game"
	"github.com/fleetdm/wordgame/internal/store"
	"github.com/fleetdm/wordgame/pkg/identifier"
)

// Server holds the dependencies needed by HTTP handlers.
// All dependencies are injected via NewServer — no global state.
type Server struct {
	store *store.GameStore
	words []string
}

// NewServer creates a Server with the given dependencies.
func NewServer(store *store.GameStore, words []string) *Server {
	return &Server{store: store, words: words}
}

// NewGameResponse is the JSON response for POST /new.
type NewGameResponse struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
}

// GuessRequest is the JSON request body for POST /guess.
type GuessRequest struct {
	ID    string `json:"id"`
	Guess string `json:"guess"`
}

// GuessResponse is the JSON response for POST /guess.
// Word is only included when the game ends (win or loss).
type GuessResponse struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
	Word             string `json:"word,omitempty"`
}

// ErrorResponse is the standard error response shape.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleNewGame handles POST /new — starts a new game.
func (s *Server) HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := identifier.GenerateIdentifier()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate game ID")
		return
	}

	word := s.words[rand.IntN(len(s.words))]
	g := game.NewGame(id, word)

	s.store.Save(g)

	snap := g.Snapshot()
	writeJSON(w, http.StatusOK, NewGameResponse{
		ID:               g.ID,
		Current:          snap.Current,
		GuessesRemaining: snap.GuessesRemaining,
	})
}

// HandleGuess handles POST /guess — makes a guess in an ongoing game.
// Applies Postel's Law: trims whitespace, uppercases, then validates.
func (s *Server) HandleGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req GuessRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Postel's Law: normalise before validation
	guess := strings.TrimSpace(req.Guess)
	guess = strings.ToUpper(guess)

	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "missing game id")
		return
	}
	if guess == "" {
		writeError(w, http.StatusBadRequest, "missing guess")
		return
	}
	if len(guess) != 1 {
		writeError(w, http.StatusBadRequest, "guess must be a single character")
		return
	}

	guessRune := rune(guess[0])
	if guessRune < 'A' || guessRune > 'Z' {
		writeError(w, http.StatusBadRequest, "guess must be a single letter A-Z")
		return
	}

	g := s.store.Get(req.ID)
	if g == nil {
		writeError(w, http.StatusBadRequest, "game not found")
		return
	}

	if err := g.ApplyGuess(guessRune); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	snap := g.Snapshot()
	resp := GuessResponse{
		ID:               g.ID,
		Current:          snap.Current,
		GuessesRemaining: snap.GuessesRemaining,
	}

	// When game ends, reveal the word and clear from store
	if snap.Status == game.StatusWon || snap.Status == game.StatusLost {
		resp.Word = g.Word
		s.store.Delete(g.ID)
	}

	writeJSON(w, http.StatusOK, resp)
}

// writeJSON encodes v as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}
