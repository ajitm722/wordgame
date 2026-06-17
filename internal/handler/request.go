package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/fleetdm/wordgame/internal/game"
)

// decodeJSONBody decodes a JSON request body with strict unknown field rejection.
func decodeJSONBody(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// normaliseGuess applies Postel's Law to the guess string:
// trims surrounding whitespace and converts to uppercase.
func normaliseGuess(guess string) string {
	return strings.ToUpper(strings.TrimSpace(guess))
}

// validateGuess checks that the guess string is a single uppercase A-Z letter.
// Returns a user-facing error message suitable for API responses.
func validateGuess(guess string) error {
	if guess == "" {
		return errors.New("missing guess")
	}
	if len(guess) != 1 {
		return errors.New("guess must be a single character")
	}
	if !game.LetterRegex.MatchString(guess) {
		return errors.New("guess must be a single letter A-Z")
	}
	return nil
}
