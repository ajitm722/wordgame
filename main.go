package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	serverAddress = "localhost:1337"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	words, err := loadWords("words.txt")
	if err != nil {
		log.Fatal(err)
	}

	// TODO use words in your implementation
	_ = words

	log.Printf("Starting server on http://%s", serverAddress)
	if err := http.ListenAndServe(serverAddress, nil); err != nil {
		log.Fatal(err)
	}
}
