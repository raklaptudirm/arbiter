package match

import (
	"math/rand"
	"os"
	"strings"
)

// NewBook opens the opening book with the given configuration.
func NewBook(name string, strategy string) (*OpeningBook, error) {
	var book OpeningBook

	// Read the opening book file.
	file, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	// Only epd books are supported currently. Split the book by newline to get
	// each opening, which are all situated on separate lines.
	book.entries = strings.Split(string(file), "\n")
	for i, entry := range book.entries {
		book.entries[i] = strings.Trim(entry, "\n\r\t ")
	}
	book.strategy = strategy

	return &book, nil
}

// OpeningBook represents a complete opening book complete with an opening
// selection strategy and state.
type OpeningBook struct {
	entries  []string
	strategy string
	current  int
}

// Next makes the book select a new opening.
func (book *OpeningBook) Next() {
	switch book.strategy {
	case "random":
		book.current = rand.Int() % len(book.entries)
	default:
		book.current = (book.current + 1) % len(book.entries)
	}
}

// Current returns the currently selected opening.
func (book *OpeningBook) Current() string {
	return book.entries[book.current]
}
