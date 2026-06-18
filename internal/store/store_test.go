package store

import (
	"testing"

	"github.com/fleetdm/wordgame/internal/game"
)

// TestNewGameStore creates a store and verifies its initial state.
func TestNewGameStore(t *testing.T) {
	s := NewGameStore()
	if s == nil {
		t.Fatal("NewGameStore returned nil")
	}
	if s.Len() != 0 {
		t.Errorf("expected 0 games, got %d", s.Len())
	}
}

// TestSaveAndGet saves a game and retrieves it, verifying ID and Word match.
func TestSaveAndGet(t *testing.T) {
	s := NewGameStore()
	g := game.NewGame("test-id", "APPLE")

	s.Save(g)

	retrieved := s.Get("test-id")
	if retrieved == nil {
		t.Fatal("Get returned nil for saved game")
	}
	if retrieved.ID != g.ID {
		t.Errorf("ID = %q, want %q", retrieved.ID, g.ID)
	}
	if retrieved.Word != g.Word {
		t.Errorf("Word = %q, want %q", retrieved.Word, g.Word)
	}
}

// TestGet_NotFound returns nil when querying a non-existent game ID.
func TestGet_NotFound(t *testing.T) {
	s := NewGameStore()

	g := s.Get("nonexistent")
	if g != nil {
		t.Error("expected nil for non-existent game")
	}
}

// TestDelete removes a game and verifies it is no longer accessible.
func TestDelete(t *testing.T) {
	s := NewGameStore()
	g := game.NewGame("test-id", "APPLE")
	s.Save(g)

	if s.Len() != 1 {
		t.Errorf("expected 1 game, got %d", s.Len())
	}

	s.Delete("test-id")

	if s.Len() != 0 {
		t.Errorf("expected 0 games after delete, got %d", s.Len())
	}
	if s.Get("test-id") != nil {
		t.Error("Get should return nil after delete")
	}
}

// TestDelete_NotFound does not panic when deleting a non-existent game.
func TestDelete_NotFound(t *testing.T) {
	s := NewGameStore()
	// Should not panic
	s.Delete("nonexistent")
	if s.Len() != 0 {
		t.Errorf("expected 0 games, got %d", s.Len())
	}
}

// TestSave_Overwrite replaces an existing game when saving with the same ID.
func TestSave_Overwrite(t *testing.T) {
	s := NewGameStore()
	g1 := game.NewGame("same-id", "APPLE")
	g2 := game.NewGame("same-id", "ORANGE")

	s.Save(g1)
	s.Save(g2)

	retrieved := s.Get("same-id")
	if retrieved.Word != "ORANGE" {
		t.Errorf("Word = %q, want %q (should be overwritten)", retrieved.Word, "ORANGE")
	}
}

// TestLen tracks the count of games across save and delete operations.
func TestLen(t *testing.T) {
	s := NewGameStore()

	if s.Len() != 0 {
		t.Errorf("Len = %d, want 0", s.Len())
	}

	s.Save(game.NewGame("id1", "APPLE"))
	if s.Len() != 1 {
		t.Errorf("Len = %d, want 1", s.Len())
	}

	s.Save(game.NewGame("id2", "ORANGE"))
	if s.Len() != 2 {
		t.Errorf("Len = %d, want 2", s.Len())
	}

	s.Delete("id1")
	if s.Len() != 1 {
		t.Errorf("Len = %d, want 1", s.Len())
	}
}
