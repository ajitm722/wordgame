// Package identifier provides UUID v4 generation for game IDs.
package identifier

import (
	"fmt"

	"github.com/google/uuid"
)

// newUUID is the function used to generate UUIDs.
// Defaults to uuid.NewRandom. Swappable in tests to cover error paths.
var newUUID = uuid.NewRandom

// GenerateIdentifier generates a new random UUID v4 string.
// Returns the UUID string or an error if generation fails.
func GenerateIdentifier() (string, error) {
	id, err := newUUID()
	if err != nil {
		return "", fmt.Errorf("generate game ID: %w", err)
	}
	return id.String(), nil
}
