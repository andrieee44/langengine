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

// Accept consumes the next rune if it is found in the given string.
func (lrd *Reader) Accept(match string) {
	lrd.AcceptFunc(func(char rune) bool {
		return strings.ContainsRune(match, char)
	})
}

// AcceptFunc consumes the next rune if fn(lrd.Next()) returns true.
func (lrd *Reader) AcceptFunc(fn func(rune) bool) {
	var char rune

	char = lrd.Next()

	if char != EOF && !fn(char) {
		lrd.Backup(1)
	}
}

// AcceptRun consumes consecutive runes while they are found in the
// given string.
func (lrd *Reader) AcceptRun(match string) {
	lrd.AcceptRunFunc(func(char rune) bool {
		return strings.ContainsRune(match, char)
	})
}

// AcceptRunFunc consumes consecutive runes while fn(lrd.Next())
// returns true.
func (lrd *Reader) AcceptRunFunc(fn func(rune) bool) {
	var char rune

	for {
		char = lrd.Next()

		if char == EOF {
			return
		}

		if !fn(char) {
			lrd.Backup(1)

			return
		}
	}
}

// Until consumes runes until EOF or until a rune is found in the
// given string.
func (lrd *Reader) Until(match string) {
	lrd.UntilFunc(func(char rune) bool {
		return strings.ContainsRune(match, char)
	})
}

// UntilFunc consumes runes until EOF or until fn(r.Next()) returns true.
func (lrd *Reader) UntilFunc(fn func(rune) bool) {
	lrd.AcceptRunFunc(func(char rune) bool {
		return !fn(char)
	})
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

// Emit returns the sequence of runes accumulated by successive calls
// to Next since the last call to Ignore or Emit, provided as a string
// along with the starting Position of that token.
func (lrd *Reader) Emit() (string, Position) {
	var (
		token string
		pos   Position
	)

	token = string(lrd.buf[lrd.start:lrd.current])
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
