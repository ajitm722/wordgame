// Command wordgame starts the word-guessing game HTTP server.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/fleetdm/wordgame/internal/handler"
	"github.com/fleetdm/wordgame/internal/store"
	"github.com/fleetdm/wordgame/pkg/words"
)

func main() {
	// 1. Load words from file — cmd/ opens the file, pkg/ reads from io.Reader
	f, err := os.Open("words.txt")
	if err != nil {
		log.Fatalf("open words.txt: %v", err)
	}
	defer f.Close()

	wordList, err := words.LoadWords(f)
	if err != nil {
		log.Fatalf("load words: %v", err)
	}
	log.Printf("loaded %d words", len(wordList))

	// 2. Create in-memory game store
	gameStore := store.NewGameStore()

	// 3. Create HTTP handler server with dependencies injected
	srv := handler.NewServer(gameStore, wordList)

	// 4. Register routes with gorilla/mux
	r := mux.NewRouter()
	r.HandleFunc("/new", srv.HandleNewGame).Methods(http.MethodPost)
	r.HandleFunc("/guess", srv.HandleGuess).Methods(http.MethodPost)

	// 5. Start listening
	addr := "localhost:" + port()
	log.Printf("starting server on http://%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

// port returns the listen port from the PORT environment variable,
// falling back to "1337" for local development.
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "1337"
}
