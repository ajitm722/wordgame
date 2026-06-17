// Package store provides a thread-safe in-memory store for Game instances.
package store

import (
	"sync"

	"github.com/fleetdm/wordgame/internal/game"
)

// GameStore provides thread-safe CRUD operations for game.Game instances.
// All methods are safe for concurrent use across goroutines.
type GameStore struct {
	mu    sync.RWMutex
	games map[string]*game.Game
}

// NewGameStore creates an empty GameStore ready for use.
func NewGameStore() *GameStore {
	return &GameStore{
		games: make(map[string]*game.Game),
	}
}

// Save stores a game in the store, keyed by its ID.
func (s *GameStore) Save(g *game.Game) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[g.ID] = g
}

// Get retrieves a game by its ID. Returns nil if not found.
func (s *GameStore) Get(id string) *game.Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.games[id]
}

// Delete removes a game from the store by its ID.
// It is safe to call Delete on a non-existent ID.
func (s *GameStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.games, id)
}

// Len returns the number of games currently stored.
func (s *GameStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.games)
}
