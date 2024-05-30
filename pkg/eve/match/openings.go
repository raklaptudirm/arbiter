package match

import (
	"math/rand"
	"os"
	"strings"
)

func NewBook(name string, strategy string) (*OpeningBook, error) {
	var book OpeningBook
	file, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	book.entries = strings.Split(string(file), "\n")
	for i, entry := range book.entries {
		book.entries[i] = strings.Trim(entry, "\n\r\t ")
	}
	book.strategy = strategy

	return &book, nil
}

type OpeningBook struct {
	entries  []string
	strategy string
	current  int
}

func (book *OpeningBook) Next() {
	switch book.strategy {
	case "random":
		book.current = rand.Int() % len(book.entries)
	default:
		book.current = (book.current + 1) % len(book.entries)
	}
}

func (book *OpeningBook) Current() string {
	return book.entries[book.current]
}
