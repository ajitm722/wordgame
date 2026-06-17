package handler

// NewGameResponse is the JSON response for POST /new.
type NewGameResponse struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
}

// GuessRequest is the JSON request body for POST /guess.
type GuessRequest struct {
	ID    string `json:"id"`
	Guess string `json:"guess"`
}

// GuessResponse is the JSON response for POST /guess.
// Word is only included when the game ends (win or loss).
type GuessResponse struct {
	ID               string `json:"id"`
	Current          string `json:"current"`
	GuessesRemaining int    `json:"guesses_remaining"`
	Word             string `json:"word,omitempty"`
}

// ErrorResponse is the standard error response shape.
type ErrorResponse struct {
	Error string `json:"error"`
}
