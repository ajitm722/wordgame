package main

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// generateIdentifier generates a new random UUID used to identify an
// in-progress game.
func generateIdentifier() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "generate game ID")
	}

	return id.String(), nil
}
