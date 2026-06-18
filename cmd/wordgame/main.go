// Command wordgame starts the word-guessing game HTTP server.
package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/fleetdm/wordgame/internal/handler"
	"github.com/fleetdm/wordgame/internal/store"
	"github.com/fleetdm/wordgame/pkg/words"
)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

// NewRootCommand constructs the CLI with --port flag and auto-generated --help.
func NewRootCommand() *cobra.Command {
	var port string

	defaultPort := "1337"
	if p := os.Getenv("PORT"); p != "" {
		defaultPort = p
	}

	cmd := &cobra.Command{
		Use:   "wordgame",
		Short: "Starts the word-guessing game HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.OutOrStderr(), port)
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", defaultPort, "Listen port for the HTTP server")
	return cmd
}

func runServer(stderr io.Writer, port string) error {
	logger := log.New(stderr, "", log.LstdFlags)

	// Load word list from filesystem
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

	// Register routes — single source of truth
	r := mux.NewRouter()
	registerRoutes(r, srv)

	// Start HTTP server
	addr := "localhost:" + port
	logger.Printf("starting server on http://%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		return err
	}
	return nil
}

// registerRoutes adds all HTTP routes to the given router.
// Shared by runServer() and smoke tests so there is a single source of truth.
func registerRoutes(r *mux.Router, srv *handler.Server) {
	r.HandleFunc("/new", srv.HandleNewGame).Methods(http.MethodPost)
	r.HandleFunc("/guess", srv.HandleGuess).Methods(http.MethodPost)
}
