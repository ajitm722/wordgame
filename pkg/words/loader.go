// Package words provides a word dictionary loader.
// Words are loaded from an io.Reader and filtered to contain
// only uppercase ASCII letters (A-Z).
//
// Thanks to https://github.com/dwyl/english-words for the word list.
package words

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// wordRegex matches lines containing only uppercase A-Z characters.
var wordRegex = regexp.MustCompile(`^[A-Z]+$`)

// LoadWords reads words from r, normalises them to uppercase,
// and filters to those containing only A-Z characters.
//
// The caller is responsible for opening/closing any underlying file.
// Passing an io.Reader (rather than a file path) decouples the loader
// from the filesystem, making it trivial to test with strings.NewReader.
func LoadWords(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)

	var words []string
	for scanner.Scan() {
		word := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if wordRegex.MatchString(word) {
			words = append(words, word)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan words: %w", err)
	}

	return words, nil
}
