package match

import (
	"os"
	"strings"
	"time"
)

type OpeningConfig struct {
	File string
	// Format string // only EPD opening files supported.
	Order string
	Start int
}

// NewBook opens the opening book with the given configuration.
func NewBook(config OpeningConfig) (*OpeningBook, error) {
	var book OpeningBook

	book.prng.Seed(config.Start)
	if config.Start == 0 {
		book.prng.Seed(int(time.Now().UnixMilli()))
	}

	// Read the opening book file.
	file, err := os.ReadFile(config.File)
	if err != nil {
		return nil, err
	}

	// Only epd books are supported currently. Split the book by newline to get
	// each opening, which are all situated on separate lines.
	book.entries = strings.Split(string(file), "\n")
	for i, entry := range book.entries {
		book.entries[i] = strings.Trim(entry, "\n\r\t ")
	}
	book.OpeningConfig = config

	return &book, nil
}

// OpeningBook represents a complete opening book complete with an opening
// selection strategy and state.
type OpeningBook struct {
	OpeningConfig
	prng    prng
	entries []string
}

// Next makes the book select a new opening.
func (book *OpeningBook) Next() {
	switch book.Order {
	case "random":
		book.Start = int(book.prng.Uint64() % uint64(len(book.entries)))
	default:
		book.Start = (book.Start + 1) % len(book.entries)
	}
}

// Current returns the currently selected opening.
func (book *OpeningBook) Current() string {
	return book.entries[book.Start]
}

// xorshift64star Pseudo-Random Number Generator
// This struct is based on original code written and dedicated
// to the public domain by Sebastiano Vigna (2014).
// It has the following characteristics:
//
//   - Outputs 64-bit numbers
//   - Passes Dieharder and SmallCrush test batteries
//   - Does not require warm-up, no zeroland to escape
//   - Internal state is a single 64-bit integer
//   - Period is 2^64 - 1
//   - Speed: 1.60 ns/call (Core i7 @3.40GHz)
//
// For further analysis see
//
//	<http://vigna.di.unimi.it/ftp/papers/xorshift.pdf>
type prng struct {
	seed uint64
}

// Seed seeds the pseudo-random number generator with the given uint.
func (p *prng) Seed(s int) {
	p.seed = uint64(s)
}

// Uint64 generates a new pseudo-random uint64.
func (p *prng) Uint64() uint64 {
	// linear feedback shifts
	p.seed ^= p.seed >> 12
	p.seed ^= p.seed << 25
	p.seed ^= p.seed >> 27

	// scramble result with non-linear function
	return p.seed * 2685821657736338717
}
