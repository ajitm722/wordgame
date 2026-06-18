// Package game implements the core word-guessing game logic.
// It is pure business logic with no I/O or HTTP concerns.
package game

import (
	"errors"
	"regexp"
	"strings"
	"sync"
)

// Sentinel errors for game operations.
var (
	ErrGameCompleted = errors.New("game already completed")
	ErrInvalidGuess  = errors.New("guess must be a single A-Z character")
)

// LetterRegex validates a single uppercase A-Z character.
// Exported so that the handler package uses the same regex.
var LetterRegex = regexp.MustCompile(`^[A-Z]$`)

// MaxGuesses is the number of wrong guesses allowed before the game is lost.
const MaxGuesses = 6

// Status represents the current state of a game.
type Status int

const (
	StatusInProgress Status = iota
	StatusWon
	StatusLost
)

// Game represents the complete state of a word-guessing game session.
type Game struct {
	ID   string // UUID v4 identifier
	Word string // The chosen word (uppercase, e.g. "APPLE")
	State       // Embedded — promoted fields: Current, GuessesRemaining, Status

	mu sync.Mutex // Protects all fields from concurrent access
}

// State holds a thread-safe snapshot of game state for external readers.
type State struct {
	Current          string
	GuessesRemaining int
	Status           Status
}

// NewGame creates a new game with the given ID and word.
// The initial board state is all underscores, one per character.
// Initial guesses remaining is MaxGuesses.
func NewGame(id, word string) *Game {
	return &Game{
		ID:   id,
		Word: word,
		State: State{
			Current:          strings.Repeat("_", len(word)),
			GuessesRemaining: MaxGuesses,
			Status:           StatusInProgress,
		},
	}
}

// ApplyGuess processes a single letter guess.
// It orchestrates validation, board update, and win/loss detection.
//
// Every guess is treated as a normal operation — no duplicate tracking.
// If the letter is in the word, all occurrences are revealed.
// If not, guesses_remaining is decremented regardless of whether
// the same wrong letter was guessed before.
//
// The guess rune MUST already be validated as A-Z. Normalisation happens in the handler.
func (g *Game) ApplyGuess(guess rune) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if err := g.validateInProgress(); err != nil {
		return err
	}
	if err := g.validateRune(guess); err != nil {
		return err
	}
	if g.isCorrectGuess(guess) {
		g.applyCorrectGuess(guess)
	} else {
		g.applyWrongGuess()
	}
	return nil
}

// validateInProgress returns ErrGameCompleted if the game has already ended.
func (g *Game) validateInProgress() error {
	if g.Status != StatusInProgress {
		return ErrGameCompleted
	}
	return nil
}

// validateRune returns ErrInvalidGuess if the rune is not uppercase A-Z.
func (g *Game) validateRune(guess rune) error {
	if !LetterRegex.MatchString(string(guess)) {
		return ErrInvalidGuess
	}
	return nil
}

// isCorrectGuess returns true if the guessed letter appears in the word.
func (g *Game) isCorrectGuess(guess rune) bool {
	return strings.ContainsRune(g.Word, guess)
}

// applyCorrectGuess reveals all occurrences of the guessed letter on the
// board and checks if the game has been won.
func (g *Game) applyCorrectGuess(guess rune) {
	runes := []rune(g.Current)
	wordRunes := []rune(g.Word)
	for i, ch := range wordRunes {
		if ch == guess {
			runes[i] = guess
		}
	}
	g.Current = string(runes)

	// Check win: all letters revealed
	if g.Current == g.Word {
		g.Status = StatusWon
	}
}

// applyWrongGuess decrements the remaining guesses and checks if the
// game has been lost.
func (g *Game) applyWrongGuess() {
	g.GuessesRemaining--

	if g.GuessesRemaining <= 0 {
		g.GuessesRemaining = 0
		g.Status = StatusLost
	}
}

// Snapshot returns a thread-safe copy of the game state for reading
// after ApplyGuess has been called. This avoids data races when the
// handler reads Current and GuessesRemaining for the HTTP response.
func (g *Game) Snapshot() State {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.State
}
