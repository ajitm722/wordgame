// Package handler implements HTTP handlers for the word-guessing game API.
// It applies Postel's Law: be liberal in what you accept, conservative in what you send.
package handler

import (
	"errors"
	"math/rand/v2"
	"net/http"

	"github.com/fleetdm/wordgame/internal/game"
	"github.com/fleetdm/wordgame/internal/store"
	"github.com/fleetdm/wordgame/pkg/identifier"
)

// generateID is the function used to generate game identifiers.
// Defaults to identifier.GenerateIdentifier. Swappable in tests to cover error paths.
var generateID = identifier.GenerateIdentifier

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

// pickWord randomly selects a word from the loaded word list.
func (s *Server) pickWord() string {
	return s.words[rand.IntN(len(s.words))]
}

// HandleNewGame handles POST /new — starts a new game.
func (s *Server) HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := generateID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate game ID")
		return
	}

	word := s.pickWord()
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
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "missing game id")
		return
	}

	// Postel's Law: normalise before validation
	guess := normaliseGuess(req.Guess)
	if err := validateGuess(guess); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	g := s.store.Get(req.ID)
	if g == nil {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}

	if err := g.ApplyGuess(rune(guess[0])); err != nil {
		switch {
		case errors.Is(err, game.ErrGameCompleted):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, game.ErrInvalidGuess):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
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
