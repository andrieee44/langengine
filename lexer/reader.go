package lexer

import (
	"io"
	"strings"
	"unicode/utf8"
)

// Position represents the location of a token in the input stream.
// It tracks both the line and column numbers, with lines incremented
// on newlines and columns incremented on each rune within a line.
type Position struct {
	// Line is the line number where the token begins.
	Line int

	// Column is the column number within the line where the token begins.
	Column int
}

// Reader provides the core lexing primitives over an io.Reader.
// It manages buffered input, position tracking, and token history,
// exposing methods such as Next, Backup, Peek, Emit, and Ignore.
// A new Reader is constructed with NewReader to set up the lexer state.
type Reader struct {
	buf                  []byte
	history              []snapshot
	rd                   io.Reader
	err                  error
	startPos, currentPos Position
	head                 int
	start, current       int
}

type snapshot struct {
	currentPos Position
	current    int
}

const (
	// EOF is the sentinel rune used to indicate end of input.
	// It is returned by Reader methods such as Next when no more
	// characters are available from the underlying source.
	EOF rune = 0

	readSize    = 4096
	initBufSize = readSize * 2
)

// NewReader constructs and returns a new Reader bound to the given io.Reader.
// The Reader is initialized with empty state and becomes ready for lexing
// once input is consumed through calls such as Next.
func NewReader(rd io.Reader) *Reader {
	var startPos Position

	startPos = Position{
		Line:   1,
		Column: 1,
	}

	return &Reader{
		rd:         rd,
		startPos:   startPos,
		currentPos: startPos,
	}
}

// AcceptSeq consumes runes matching the exact sequence of the given
// string. It advances the reader rune by rune and checks whether each
// rune matches in order.
//
// Returns true if the entire sequence was successfully consumed.
// Returns false if EOF is reached or a mismatch occurs (in which case
// the reader position is restored via Backup).
func (lrd *Reader) AcceptSeq(match string) bool {
	var (
		runes []rune
		char  rune
		count int
	)

	runes = []rune(match)

	for _, char = range runes {
		if lrd.Next() != char {
			break
		}

		count++

	}

	if count != len(runes) {
		lrd.Backup(count + 1)

		return false
	}

	return true
}

// Accept consumes the next rune if it is found in the given string.
// It advances the reader by one rune and checks whether that rune
// exists within the provided match string.
//
// Returns true if the next rune was successfully consumed (i.e., it
// was found in match). Returns false if the next rune was EOF or not
// present in match (in which case the reader position is restored via
// Backup).
func (lrd *Reader) Accept(match string) bool {
	return lrd.AcceptFunc(func(char rune) bool {
		return strings.ContainsRune(match, char)
	})
}

// AcceptFunc consumes the next rune if the provided predicate function
// returns true. It advances the reader by one rune and applies fn to it.
//
// Returns true if the next rune was successfully consumed (i.e., fn
// returned true). Returns false if the next rune was EOF or if fn
// returned false (in which case the reader position is restored via
// Backup).
func (lrd *Reader) AcceptFunc(fn func(rune) bool) bool {
	var char rune

	char = lrd.Next()

	if char == EOF {
		return false
	}

	if !fn(char) {
		lrd.Backup(1)

		return false
	}

	return true
}

// AcceptRun consumes consecutive runes while they are found in the
// given string. It advances the reader rune by rune and checks whether
// each rune exists within the provided match string.
//
// Returns the number of runes successfully consumed. Stops and returns
// when the next rune is EOF or not present in match (in which case the
// reader position is restored via Backup).
func (lrd *Reader) AcceptRun(match string) int {
	return lrd.AcceptRunFunc(func(char rune) bool {
		return strings.ContainsRune(match, char)
	})
}

// AcceptRunFunc consumes consecutive runes while the provided predicate
// function returns true. It advances the reader rune by rune and applies
// fn to each.
//
// Returns the number of runes successfully consumed. Stops and returns
// when the next rune is EOF or when fn returns false (in which case the
// reader position is restored via Backup).
func (lrd *Reader) AcceptRunFunc(fn func(rune) bool) int {
	var (
		char  rune
		count int
	)

	for {
		char = lrd.Next()

		if char == EOF {
			return count
		}

		if !fn(char) {
			lrd.Backup(1)

			return count
		}

		count++
	}
}

// Until consumes runes until EOF or until a rune is found in the
// given string. It advances the reader rune by rune and checks whether
// each rune exists within the provided match string.
//
// Returns the number of runes successfully consumed. Stops and returns
// when the next rune is EOF or when a rune is found in match (in which
// case the reader position is restored via Backup).
func (lrd *Reader) Until(match string) int {
	return lrd.UntilFunc(func(char rune) bool {
		return strings.ContainsRune(match, char)
	})
}

// UntilFunc consumes runes until EOF or until the provided predicate
// function returns true. It advances the reader rune by rune and applies
// fn to each, stopping once fn returns true.
//
// Returns the number of runes successfully consumed. Stops and returns
// when the next rune is EOF or when fn returns true (in which case the
// reader position is restored via Backup).
func (lrd *Reader) UntilFunc(fn func(rune) bool) int {
	return lrd.AcceptRunFunc(func(char rune) bool {
		return !fn(char)
	})
}

// UntilSeq consumes runes until EOF or until the exact sequence of the
// given string is found. It advances the reader rune by rune until the
// first rune of match is encountered, then checks whether the remainder
// of the sequence follows.
//
// Returns the number of runes successfully consumed before the start of
// the matched sequence. Stops and returns when the next rune is EOF or
// when the full sequence is found (in which case the reader position is
// restored via Backup).
func (lrd *Reader) UntilSeq(match string) int {
	var (
		runes []rune
		count int
	)

	runes = []rune(match)
	count = lrd.UntilSeqInclusive(match)
	lrd.Backup(len(runes))

	return count - len(runes)
}

// UntilSeq consumes runes until EOF or until the exact sequence of the
// given string is found. It advances the reader rune by rune until the
// first rune of match is encountered, then checks whether the remainder
// of the sequence follows.
//
// Returns the number of runes successfully consumed before the start of
// the matched sequence. Stops and returns when the next rune is EOF or
// when the full sequence is found (in which case the reader position is
// restored via Backup).
func (lrd *Reader) UntilSeqInclusive(match string) int {
	var (
		runes []rune
		char  rune
		count int
	)

	runes = []rune(match)
	if len(runes) == 0 {
		return 0
	}

	for {
		count += lrd.Until(string(runes[0]))

		char = lrd.Next()
		if char == EOF {
			return count
		}

		count++

		if !lrd.AcceptSeq(string(runes[1:])) {
			continue
		}

		return count + len(runes) - 1
	}
}

// Next returns the next rune from the input stream.
// When the end of input is reached, Next returns EOF.
// Don't forget to check Err when encountering EOF.
func (lrd *Reader) Next() rune {
	var (
		char rune
		size int
	)

	lrd.fill()

	if lrd.head-lrd.current <= 0 {
		return EOF
	}

	lrd.history = append(lrd.history, snapshot{
		current:    lrd.current,
		currentPos: lrd.currentPos,
	})

	char, size = utf8.DecodeRune(lrd.buf[lrd.current:lrd.head])
	lrd.current += size

	lrd.currentPos.Column++
	if char == '\n' {
		lrd.currentPos.Line++
		lrd.currentPos.Column = 1
	}

	return char
}

// Peek returns the next rune from the input stream without advancing
// the Reader’s position. Unlike Next, it does not consume the rune.
func (lrd *Reader) Peek() rune {
	var char rune

	char = lrd.Next()
	lrd.Backup(1)

	return char
}

// Backup rewinds the Reader’s position by up to n runes, restoring
// previously consumed input. Supplying a value of n larger than the
// available history is safe: Backup will stop automatically at the
// starting rune without panicking.
func (lrd *Reader) Backup(n int) {
	var snap snapshot

	for range n {
		if len(lrd.history) == 0 {
			return
		}

		snap = lrd.history[len(lrd.history)-1]
		lrd.history = lrd.history[:len(lrd.history)-1]

		lrd.current = snap.current
		lrd.currentPos = snap.currentPos
	}
}

// Ignore discards the runes accumulated by successive calls to Next
// since the last call to Ignore or Emit, resetting the start position
// for the next token.
func (lrd *Reader) Ignore() {
	lrd.start = lrd.current
	lrd.startPos = lrd.currentPos
	lrd.history = lrd.history[:0]
}

// PeekToken returns the sequence of runes accumulated by successive
// calls to Next since the last call to Ignore or Emit, without
// consuming them. Unlike Emit, it does not advance the Reader’s
// position or reset the token boundaries.
func (lrd *Reader) PeekToken() string {
	return string(lrd.buf[lrd.start:lrd.current])
}

// Emit returns the sequence of runes accumulated by successive calls
// to Next since the last call to Ignore or Emit, provided as a string
// along with the starting Position of that token.
func (lrd *Reader) Emit() (string, Position) {
	var (
		token string
		pos   Position
	)

	token = lrd.PeekToken()
	pos = lrd.startPos

	lrd.Ignore()

	return token, pos
}

// Err returns the first error encountered from the underlying io.Reader,
// including io.EOF. This should be checked after Next returns EOF to
// distinguish between a clean end of input and other error conditions.
// A successful read sequence is indicated when Next returns EOF and
// Err returns io.EOF. In cases where EOF is returned with a nil error,
// the underlying reader may not yet be ready to provide data, and the
// client can decide how to proceed.
func (lrd *Reader) Err() error {
	return lrd.err
}

func (lrd *Reader) fill() {
	var (
		newBuf []byte
		n      int
		err    error
	)

	if lrd.buf == nil {
		lrd.buf = make([]byte, initBufSize)
	}

	switch {
	case lrd.err == io.EOF || lrd.head-lrd.current >= utf8.UTFMax:
		return
	case len(lrd.buf)-lrd.head >= readSize:
		// Do nothing
	case lrd.current-lrd.start >= len(lrd.buf)-readSize:
		newBuf = make([]byte, len(lrd.buf)*2)
		copy(newBuf, lrd.buf)
		lrd.buf = newBuf
	default:
		lrd.head -= lrd.start
		lrd.current -= lrd.start
		copy(lrd.buf, lrd.buf[lrd.start:])
		lrd.start = 0
	}

	n, err = lrd.rd.Read(lrd.buf[lrd.head : lrd.head+readSize])
	if n < 0 || n > readSize {
		panic("langengine/lexer: bogus io.Reader")
	}

	lrd.head += n

	if lrd.err == nil && err != nil {
		lrd.err = err
	}
}
