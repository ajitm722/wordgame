// Command wordgame starts the word-guessing game HTTP server.
package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/fleetdm/wordgame/internal/handler"
	"github.com/fleetdm/wordgame/internal/store"
	"github.com/fleetdm/wordgame/pkg/words"
)

func main() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		log.Fatal(err)
	}
}

// run contains the server startup logic, extracted from main() for testability.
// args is reserved for future CLI flag parsing (e.g. -port, -words-file).
// stdout is reserved for future structured output (e.g. JSON machine-readable status).
// stderr is wired into a custom logger so tests can capture startup messages.
func run(args []string, stdout, stderr io.Writer) error {
	logger := log.New(stderr, "", log.LstdFlags)

	f, err := os.Open("words.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	wordList, err := words.LoadWords(f)
	if err != nil {
		return err
	}
	logger.Printf("loaded %d words from words.txt", len(wordList))

	// Create in-memory game store
	gameStore := store.NewGameStore()

	// Create HTTP handler server with dependencies injected
	srv := handler.NewServer(gameStore, wordList)

	// Register routes with gorilla/mux
	r := mux.NewRouter()
	r.HandleFunc("/new", srv.HandleNewGame).Methods(http.MethodPost)
	r.HandleFunc("/guess", srv.HandleGuess).Methods(http.MethodPost)

	// Start listening
	addr := "localhost:" + port()
	logger.Printf("starting server on http://%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		return err
	}
	return nil
}

// port returns the listen port from the PORT environment variable,
// falling back to "1337" for local development.
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "1337"
}
