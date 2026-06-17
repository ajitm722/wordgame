package game

import (
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
	g.ApplyGuess('Z')
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
	g.ApplyGuess('P')
	g.ApplyGuess('L')
	g.ApplyGuess('E')

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

func TestApplyGuess_RevealsAllOnWin(t *testing.T) {
	g := NewGame("test-id", "AB")
	g.Current = "A_"

	err := g.ApplyGuess('B')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Current != "AB" {
		t.Errorf("Current = %q, want %q", g.Current, "AB")
	}
	if g.Status != StatusWon {
		t.Errorf("Status should be Won")
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
