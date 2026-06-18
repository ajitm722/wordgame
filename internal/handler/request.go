package handler

import (
	"encoding/json"
	"net/http"
	"strings"
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
