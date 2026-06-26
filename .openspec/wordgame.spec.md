# WordGame Behavior Specification

This specification defines the strict behavioral contract for the WordGame REST API. 
All AI-generated code and tests MUST adhere to these scenarios.

## Feature: Game Lifecycle & Guessing

### Scenario: Starting a new game
**Given** the server is running and `words.txt` is loaded
**When** a client makes a `POST` request to `/new`
**Then** the server must generate a new UUID v4
**And** select a random word from the dictionary
**And** return a `200 OK` response
**And** the response body must contain `id`, `current` (all underscores), and `guesses_remaining` (6).

### Scenario: Guessing a correct letter
**Given** an active game exists with the word "APPLE"
**When** a client sends a `POST` request to `/guess` with guess "P"
**Then** the server must return a `200 OK`
**And** the `current` state must update to "_ P P _ _"
**And** the `guesses_remaining` must NOT decrease.

### Scenario: Guessing an incorrect letter
**Given** an active game exists with the word "APPLE" and 6 guesses remaining
**When** a client sends a `POST` request to `/guess` with guess "Z"
**Then** the server must return a `200 OK`
**And** the `current` state must NOT change
**And** the `guesses_remaining` must decrease by 1 to 5.

### Scenario: Winning the game
**Given** an active game exists where only one letter is missing (e.g., "_ P P L E")
**When** a client guesses the final correct letter ("A")
**Then** the server must return a `200 OK`
**And** the `current` state must reveal the full word ("A P P L E")
**And** the response MUST include the actual `word` field
**And** the game must be immediately deleted from the store.

### Scenario: Losing the game
**Given** an active game exists with 1 guess remaining
**When** a client guesses an incorrect letter
**Then** the server must return a `200 OK`
**And** the `guesses_remaining` must reach 0
**And** the response MUST include the actual `word` field
**And** the game must be immediately deleted from the store.

### Scenario: Invalid or malformed guess (Postel's Law)
**Given** an active game exists
**When** a client guesses " a " (lowercase with whitespace)
**Then** the server must normalize it to "A" and process it normally.
**Except When** the guess is multiple letters (e.g., "AB") or non-alphabetic ("1")
**Then** the server must return a `422 Unprocessable Entity`.

### Scenario: Guessing on a completed or non-existent game
**Given** a game that has already been won, lost, or never existed
**When** a client makes a `POST` request to `/guess` for that game ID
**Then** the server must return a `404 Not Found`.
