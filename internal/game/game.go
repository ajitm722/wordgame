// Package game implements the core word-guessing game logic.
// It is pure business logic with no I/O or HTTP concerns.
package game

import (
	"fmt"
	"strings"
	"sync"
)

// Status represents the current state of a game.
type Status int

const (
	StatusInProgress Status = iota
	StatusWon
	StatusLost
)

// Game represents the complete state of a word-guessing game session.
type Game struct {
	ID               string // UUID v4 identifier
	Word             string // The chosen word (uppercase, e.g. "APPLE")
	Current          string // Board state with underscores (e.g. "_PP__")
	GuessesRemaining int    // Starts at 6, counts down on wrong guesses
	Status           Status // InProgress, Won, or Lost

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
// Initial guesses remaining is 6.
func NewGame(id, word string) *Game {
	return &Game{
		ID:               id,
		Word:             word,
		Current:          strings.Repeat("_", len(word)),
		GuessesRemaining: 6,
		Status:           StatusInProgress,
	}
}

// ApplyGuess processes a single letter guess.
// Every guess is treated as a normal operation — no duplicate tracking.
// If the letter is in the word, all occurrences are revealed.
// If not, guesses_remaining is decremented regardless of whether
// the same wrong letter was guessed before.
//
// The guess rune MUST already be validated as A-Z. Normalisation happens in the handler.
//
// Returns an error if:
//   - The game has already been won or lost
//   - The guess rune is not A-Z (defensive check)
func (g *Game) ApplyGuess(guess rune) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusInProgress {
		return fmt.Errorf("game already completed")
	}

	if guess < 'A' || guess > 'Z' {
		return fmt.Errorf("guess must be a single A-Z character")
	}

	if strings.ContainsRune(g.Word, guess) {
		// Correct guess — reveal all occurrences
		runes := []rune(g.Current)
		wordRunes := []rune(g.Word)
		for i, ch := range wordRunes {
			if ch == guess {
				runes[i] = guess
			}
		}
		g.Current = string(runes)

		// Check win
		if g.Current == g.Word {
			g.Status = StatusWon
		}
	} else {
		// Wrong guess — always decrement, even if guessed before
		g.GuessesRemaining--

		// Check loss
		if g.GuessesRemaining <= 0 {
			g.GuessesRemaining = 0
			g.Status = StatusLost
		}
	}

	return nil
}

// Snapshot returns a thread-safe copy of the game state for reading
// after ApplyGuess has been called. This avoids data races when the
// handler reads Current and GuessesRemaining for the HTTP response.
func (g *Game) Snapshot() State {
	g.mu.Lock()
	defer g.mu.Unlock()
	return State{
		Current:          g.Current,
		GuessesRemaining: g.GuessesRemaining,
		Status:           g.Status,
	}
}
