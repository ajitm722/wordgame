package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Thanks to https://github.com/dwyl/english-words for the word list.

var wordsRegexp = regexp.MustCompile("^[A-Z]+$")

// loadWords loads the word dictionary from the provided file path.
func loadWords(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open word file")
	}
	defer f.Close()

	words := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Normalize and filter words.
		word := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if wordsRegexp.MatchString(word) {
			words = append(words, word)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scan words")
	}

	return words, nil
}
