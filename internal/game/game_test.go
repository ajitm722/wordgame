package game

import (
	"errors"
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame("test-id", "APPLE")

	if g.ID != "test-id" {
		t.Errorf("ID = %q, want %q", g.ID, "test-id")
	}
	if g.Word != "APPLE" {
		t.Errorf("Word = %q, want %q", g.Word, "APPLE")
	}
	if g.Current != "_____" {
		t.Errorf("Current = %q, want %q", g.Current, "_____")
	}
	if g.GuessesRemaining != 6 {
		t.Errorf("GuessesRemaining = %d, want 6", g.GuessesRemaining)
	}
	if g.Status != StatusInProgress {
		t.Errorf("Status = %d, want StatusInProgress", g.Status)
	}
}

func TestNewGame_DifferentLengths(t *testing.T) {
	tests := []struct {
		word            string
		expectedCurrent string
	}{
		{"A", "_"},
		{"AB", "__"},
		{"ABC", "___"},
		{"APPLE", "_____"},
		{"BANANA", "______"},
	}

	for _, tt := range tests {
		g := NewGame("id", tt.word)
		if g.Current != tt.expectedCurrent {
			t.Errorf("NewGame(%q): Current = %q, want %q", tt.word, g.Current, tt.expectedCurrent)
		}
	}
}

func TestApplyGuess_Correct(t *testing.T) {
	g := NewGame("test-id", "APPLE")

	err := g.ApplyGuess('P')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Current != "_PP__" {
		t.Errorf("Current = %q, want %q", g.Current, "_PP__")
	}
	if g.GuessesRemaining != 6 {
		t.Errorf("GuessesRemaining = %d, want 6", g.GuessesRemaining)
	}
	if g.Status != StatusInProgress {
		t.Errorf("Status should be InProgress")
	}
}

func TestApplyGuess_Correct_MultipleOccurrences(t *testing.T) {
	g := NewGame("test-id", "BANANA")

	err := g.ApplyGuess('A')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Current != "_A_A_A" {
		t.Errorf("Current = %q, want %q", g.Current, "_A_A_A")
	}
	if g.GuessesRemaining != 6 {
		t.Errorf("GuessesRemaining = %d, want 6", g.GuessesRemaining)
	}
}

func TestApplyGuess_Wrong(t *testing.T) {
	g := NewGame("test-id", "APPLE")

	err := g.ApplyGuess('Z')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Current != "_____" {
		t.Errorf("Current should remain unchanged")
	}
	if g.GuessesRemaining != 5 {
		t.Errorf("GuessesRemaining = %d, want 5", g.GuessesRemaining)
	}
}

func TestApplyGuess_Win(t *testing.T) {
	g := NewGame("test-id", "CAT")
	g.Current = "CA_"

	err := g.ApplyGuess('T')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Current != "CAT" {
		t.Errorf("Current = %q, want %q", g.Current, "CAT")
	}
	if g.Status != StatusWon {
		t.Errorf("Status should be Won, got %d", g.Status)
	}
}

func TestApplyGuess_Loss(t *testing.T) {
	g := NewGame("test-id", "DOG")
	g.GuessesRemaining = 1

	err := g.ApplyGuess('Q')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.GuessesRemaining != 0 {
		t.Errorf("GuessesRemaining = %d, want 0", g.GuessesRemaining)
	}
	if g.Status != StatusLost {
		t.Errorf("Status should be Lost, got %d", g.Status)
	}
}

func TestApplyGuess_AllSixWrong_Loses(t *testing.T) {
	g := NewGame("test-id", "XYZ")

	letters := []rune{'A', 'B', 'C', 'D', 'E', 'F'}
	for i, letter := range letters {
		err := g.ApplyGuess(letter)
		if err != nil {
			t.Fatalf("guess %d: unexpected error: %v", i+1, err)
		}
	}

	if g.GuessesRemaining != 0 {
		t.Errorf("GuessesRemaining = %d, want 0", g.GuessesRemaining)
	}
	if g.Status != StatusLost {
		t.Errorf("Status should be Lost")
	}
}

func TestApplyGuess_RepeatWrong_DecrementsAgain(t *testing.T) {
	g := NewGame("test-id", "APPLE")

	// First wrong guess
	_ = g.ApplyGuess('Z')
	if g.GuessesRemaining != 5 {
		t.Fatalf("after first Z: GuessesRemaining = %d, want 5", g.GuessesRemaining)
	}

	// Repeat the same wrong letter — must decrement again
	err := g.ApplyGuess('Z')
	if err != nil {
		t.Fatalf("unexpected error on repeat wrong: %v", err)
	}
	if g.GuessesRemaining != 4 {
		t.Errorf("GuessesRemaining = %d, want 4 (repeat wrong should decrement again)", g.GuessesRemaining)
	}
}

func TestApplyGuess_RepeatCorrect_RevealsAgain(t *testing.T) {
	g := NewGame("test-id", "APPLE")
	_ = g.ApplyGuess('P')
	_ = g.ApplyGuess('L')
	_ = g.ApplyGuess('E')

	// Repeat a correct letter 'P' — already revealed, no change, no penalty
	beforeGuesses := g.GuessesRemaining
	beforeCurrent := g.Current

	err := g.ApplyGuess('P')
	if err != nil {
		t.Fatalf("unexpected error on repeat correct: %v", err)
	}
	if g.GuessesRemaining != beforeGuesses {
		t.Errorf("GuessesRemaining changed from %d to %d (repeat correct should not penalise)",
			beforeGuesses, g.GuessesRemaining)
	}
	if g.Current != beforeCurrent {
		t.Errorf("Current changed from %q to %q (repeat correct should not change board)",
			beforeCurrent, g.Current)
	}
}

func TestApplyGuess_InvalidRune(t *testing.T) {
	g := NewGame("test-id", "APPLE")

	if err := g.ApplyGuess('5'); err == nil {
		t.Error("expected error for non-letter guess")
	}
	if err := g.ApplyGuess('é'); err == nil {
		t.Error("expected error for non-ASCII guess")
	}
	if err := g.ApplyGuess('a'); err == nil {
		t.Error("expected error for lowercase (should be normalised by handler)")
	}
	if err := g.ApplyGuess('Z'); err != nil {
		t.Error("unexpected error for valid uppercase letter")
	}
}

func TestApplyGuess_AlreadyCompleted(t *testing.T) {
	g := NewGame("test-id", "APPLE")
	g.Status = StatusWon

	err := g.ApplyGuess('A')
	if err == nil {
		t.Error("expected error for completed game")
	}
}

func TestApplyGuess_AlreadyLost(t *testing.T) {
	g := NewGame("test-id", "APPLE")
	g.Status = StatusLost

	err := g.ApplyGuess('A')
	if err == nil {
		t.Error("expected error for lost game")
	}
}

func TestApplyGuess_RepeatWrongLoses(t *testing.T) {
	// Lose by guessing the same wrong letter 6 times
	g := NewGame("test-id", "XYZ")

	for i := 0; i < 6; i++ {
		err := g.ApplyGuess('A')
		if err != nil {
			t.Fatalf("guess %d: unexpected error: %v", i+1, err)
		}
	}

	if g.GuessesRemaining != 0 {
		t.Errorf("GuessesRemaining = %d, want 0", g.GuessesRemaining)
	}
	if g.Status != StatusLost {
		t.Errorf("Status should be Lost")
	}
}

// --- SRP method tests ---

func TestValidateInProgress(t *testing.T) {
	t.Run("in progress returns nil", func(t *testing.T) {
		g := NewGame("id", "TEST")
		if err := g.validateInProgress(); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("already won returns error", func(t *testing.T) {
		g := NewGame("id", "TEST")
		g.Status = StatusWon
		if err := g.validateInProgress(); !errors.Is(err, ErrGameCompleted) {
			t.Errorf("expected ErrGameCompleted, got %v", err)
		}
	})

	t.Run("already lost returns error", func(t *testing.T) {
		g := NewGame("id", "TEST")
		g.Status = StatusLost
		if err := g.validateInProgress(); !errors.Is(err, ErrGameCompleted) {
			t.Errorf("expected ErrGameCompleted, got %v", err)
		}
	})
}

func TestValidateRune(t *testing.T) {
	g := NewGame("id", "TEST")

	t.Run("valid A-Z", func(t *testing.T) {
		if err := g.validateRune('M'); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("lowercase returns error", func(t *testing.T) {
		if err := g.validateRune('a'); !errors.Is(err, ErrInvalidGuess) {
			t.Errorf("expected ErrInvalidGuess, got %v", err)
		}
	})

	t.Run("digit returns error", func(t *testing.T) {
		if err := g.validateRune('5'); !errors.Is(err, ErrInvalidGuess) {
			t.Errorf("expected ErrInvalidGuess, got %v", err)
		}
	})

	t.Run("special char returns error", func(t *testing.T) {
		if err := g.validateRune('@'); !errors.Is(err, ErrInvalidGuess) {
			t.Errorf("expected ErrInvalidGuess, got %v", err)
		}
	})

	t.Run("non-ASCII returns error", func(t *testing.T) {
		if err := g.validateRune('é'); !errors.Is(err, ErrInvalidGuess) {
			t.Errorf("expected ErrInvalidGuess, got %v", err)
		}
	})
}

func TestIsCorrectGuess(t *testing.T) {
	g := NewGame("id", "BANANA")

	t.Run("letter in word", func(t *testing.T) {
		if !g.isCorrectGuess('A') {
			t.Error("expected true for 'A' in BANANA")
		}
	})

	t.Run("letter not in word", func(t *testing.T) {
		if g.isCorrectGuess('Z') {
			t.Error("expected false for 'Z' not in BANANA")
		}
	})

	t.Run("single-letter word correct", func(t *testing.T) {
		g2 := NewGame("id", "X")
		if !g2.isCorrectGuess('X') {
			t.Error("expected true for 'X' in 'X'")
		}
	})
}

func TestApplyCorrectGuess(t *testing.T) {
	t.Run("reveals letter on board", func(t *testing.T) {
		g := NewGame("id", "APPLE")
		g.applyCorrectGuess('P')
		if g.Current != "_PP__" {
			t.Errorf("Current = %q, want %q", g.Current, "_PP__")
		}
		if g.Status != StatusInProgress {
			t.Errorf("should still be in progress after partial reveal")
		}
	})

	t.Run("wins when fully revealed", func(t *testing.T) {
		g := NewGame("id", "AB")
		g.Current = "A_"
		g.applyCorrectGuess('B')
		if g.Current != "AB" {
			t.Errorf("Current = %q, want %q", g.Current, "AB")
		}
		if g.Status != StatusWon {
			t.Errorf("Status should be Won, got %d", g.Status)
		}
	})
}

func TestApplyWrongGuess(t *testing.T) {
	t.Run("decrements remaining", func(t *testing.T) {
		g := NewGame("id", "APPLE")
		g.applyWrongGuess()
		if g.GuessesRemaining != 5 {
			t.Errorf("GuessesRemaining = %d, want 5", g.GuessesRemaining)
		}
		if g.Status != StatusInProgress {
			t.Errorf("should still be in progress")
		}
	})

	t.Run("loses when zero remaining", func(t *testing.T) {
		g := NewGame("id", "APPLE")
		g.GuessesRemaining = 1
		g.applyWrongGuess()
		if g.GuessesRemaining != 0 {
			t.Errorf("GuessesRemaining = %d, want 0", g.GuessesRemaining)
		}
		if g.Status != StatusLost {
			t.Errorf("Status should be Lost, got %d", g.Status)
		}
	})

	t.Run("caps at zero", func(t *testing.T) {
		g := NewGame("id", "APPLE")
		g.GuessesRemaining = 0
		g.applyWrongGuess()
		if g.GuessesRemaining != 0 {
			t.Errorf("GuessesRemaining = %d, want 0 (should not go negative)", g.GuessesRemaining)
		}
		if g.Status != StatusLost {
			t.Errorf("Status should be Lost, got %d", g.Status)
		}
	})
}

func TestSnapshot(t *testing.T) {
	g := NewGame("test-id", "APPLE")
	_ = g.ApplyGuess('P')

	snap := g.Snapshot()
	if snap.Current != "_PP__" {
		t.Errorf("snapshot Current = %q, want %q", snap.Current, "_PP__")
	}
	if snap.GuessesRemaining != 6 {
		t.Errorf("snapshot GuessesRemaining = %d, want 6", snap.GuessesRemaining)
	}
	if snap.Status != StatusInProgress {
		t.Errorf("snapshot Status should be InProgress")
	}
}
