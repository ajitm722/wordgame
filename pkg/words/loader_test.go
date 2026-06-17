package words

import (
	"errors"
	"strings"
	"testing"
)

// errReader returns a read error on every call, used to test scanner error handling.
type errReader struct{}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestLoadWords_Basic(t *testing.T) {
	input := strings.NewReader("apple\norange\nbanana\n123abc\nhéllo\n")
	words, err := LoadWords(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"APPLE", "ORANGE", "BANANA"}
	if len(words) != len(expected) {
		t.Fatalf("expected %d words, got %d: %v", len(expected), len(words), words)
	}
	for i, w := range expected {
		if words[i] != w {
			t.Errorf("words[%d] = %q, want %q", i, words[i], w)
		}
	}
}

func TestLoadWords_Empty(t *testing.T) {
	words, err := LoadWords(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words, got %d", len(words))
	}
}

func TestLoadWords_Whitespace(t *testing.T) {
	input := strings.NewReader("  apple  \n  ORANGE  \n")
	words, err := LoadWords(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"APPLE", "ORANGE"}
	if len(words) != len(expected) {
		t.Fatalf("expected %d words, got %d", len(expected), len(words))
	}
	for i, w := range expected {
		if words[i] != w {
			t.Errorf("words[%d] = %q, want %q", i, words[i], w)
		}
	}
}

func TestLoadWords_FilterNonAlpha(t *testing.T) {
	input := strings.NewReader("hello\nh3llo\nhe'llo\nhéllo\n")
	words, err := LoadWords(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only "HELLO" should pass the A-Z filter
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d: %v", len(words), words)
	}
	if words[0] != "HELLO" {
		t.Errorf("words[0] = %q, want %q", words[0], "HELLO")
	}
}

func TestLoadWords_MixedCase(t *testing.T) {
	input := strings.NewReader("Apple\nORANGE\nBanana\n")
	words, err := LoadWords(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"APPLE", "ORANGE", "BANANA"}
	if len(words) != len(expected) {
		t.Fatalf("expected %d words, got %d", len(expected), len(words))
	}
	for i, w := range expected {
		if words[i] != w {
			t.Errorf("words[%d] = %q, want %q", i, words[i], w)
		}
	}
}

func TestLoadWords_SingleLetter(t *testing.T) {
	input := strings.NewReader("a\nb\nc\n")
	words, err := LoadWords(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"A", "B", "C"}
	if len(words) != len(expected) {
		t.Fatalf("expected %d words, got %d", len(expected), len(words))
	}
	for i, w := range expected {
		if words[i] != w {
			t.Errorf("words[%d] = %q, want %q", i, words[i], w)
		}
	}
}

// --- SRP validation tests ---

func TestIsValidWord(t *testing.T) {
	tests := []struct {
		word  string
		valid bool
	}{
		{"HELLO", true},
		{"A", true},
		{"Z", true},
		{"APPLE", true},
		{"BANANA", true},
		{"", false},
		{"HEL3O", false},
		{"HEL-LO", false},
		{"HÉLLO", false},
		{"hello", false},
		{"HELLO ", false},
		{" HELLO", false},
	}

	for _, tt := range tests {
		got := isValidWord(tt.word)
		if got != tt.valid {
			t.Errorf("isValidWord(%q) = %v, want %v", tt.word, got, tt.valid)
		}
	}
}

// TestLoadWords_ScannerError verifies that when the underlying reader
// returns an error (making bufio.Scanner.Err() non-nil), the error is
// propagated to the caller.
func TestLoadWords_ScannerError(t *testing.T) {
	words, err := LoadWords(&errReader{})
	if err == nil {
		t.Fatal("expected error from scanner, got nil")
	}
	if words != nil {
		t.Errorf("expected nil words on scanner error, got %d words", len(words))
	}
}
