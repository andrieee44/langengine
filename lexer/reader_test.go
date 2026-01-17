package lexer

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type bogusReader struct{}

func (r bogusReader) Read(p []byte) (n int, err error) {
	return -1, nil
}

func assertBuf(t *testing.T, expected, got []byte) {
	var i int

	t.Helper()

	if len(got) < len(expected) {
		t.Errorf(
			"got slice shorter than expected (len=%d, expected=%d)",
			len(got),
			len(expected),
		)
	}

	for i = 0; i < len(expected); i++ {
		if got[i] == expected[i] {
			continue
		}

		t.Errorf(
			"got %q, expected %q in index %d",
			got[i],
			expected[i],
			i,
		)
	}

	for i = len(expected); i < len(got); i++ {
		if got[i] == 0 {
			continue
		}

		t.Errorf(
			"got %q, expected %q in index %d",
			got[i],
			byte(0),
			i,
		)
	}
}

func TestReaderFill(t *testing.T) {
	t.Parallel()

	t.Run("immediateEOF", func(t *testing.T) {
		var lrd *Reader

		t.Parallel()

		lrd = NewReader(strings.NewReader(""))
		lrd.fill()

		assert.Equal(t, io.EOF, lrd.Err())
		assert.Equal(t, 0, lrd.head)
	})

	t.Run("once", func(t *testing.T) {
		var (
			testTbl [][]byte
			buf     []byte
		)

		t.Parallel()

		testTbl = [][]byte{
			[]byte("a"),
			[]byte("abcdefghijk"),
			bytes.Repeat([]byte{'A'}, readSize),
			bytes.Repeat([]byte{'B'}, initBufSize),
		}

		for _, buf = range testTbl {
			t.Run(fmt.Sprintf("%q", buf), func(t *testing.T) {
				var (
					lrd    *Reader
					bufLen int
				)

				lrd = NewReader(bytes.NewReader(buf))
				lrd.fill()

				bufLen = min(len(buf), readSize)

				assert.Equal(t, nil, lrd.Err())
				assert.Equal(t, bufLen, lrd.head)
				assert.Equal(t, initBufSize, len(lrd.buf))
				assertBuf(t, buf[:bufLen], lrd.buf)
			})
		}

	})

	t.Run("shortDoNothing", func(t *testing.T) {
		var (
			buf []byte
			lrd *Reader
		)

		t.Parallel()

		// Less than utf8.UTFMax (4)
		buf = []byte("ab")

		lrd = NewReader(bytes.NewReader(buf))
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, len(buf), lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf, lrd.buf)

		lrd.fill()

		assert.Equal(t, io.EOF, lrd.Err())
		assert.Equal(t, len(buf), lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf, lrd.buf)
	})

	t.Run("doNothing", func(t *testing.T) {
		var (
			buf []byte
			lrd *Reader
		)

		t.Parallel()

		buf = []byte("qwertyuiop")

		lrd = NewReader(bytes.NewReader(buf))
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, len(buf), lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf, lrd.buf)

		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, len(buf), lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf, lrd.buf)
	})

	t.Run("grow", func(t *testing.T) {
		var (
			buf []byte
			lrd *Reader
		)

		t.Parallel()

		buf = bytes.Repeat([]byte{'A'}, readSize*3)

		lrd = NewReader(bytes.NewReader(buf))
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, readSize, lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf[:lrd.head], lrd.buf)

		lrd.current = lrd.head
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, readSize*2, lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf[:lrd.head], lrd.buf)

		lrd.current = lrd.head
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, readSize*3, lrd.head)
		assert.Equal(t, initBufSize*2, len(lrd.buf))
		assertBuf(t, buf[:lrd.head], lrd.buf)

		lrd.current = lrd.head
		lrd.fill()

		assert.Equal(t, io.EOF, lrd.Err())
		assert.Equal(t, readSize*3, lrd.head)
		assert.Equal(t, initBufSize*2, len(lrd.buf))
		assertBuf(t, buf[:lrd.head], lrd.buf)
	})

	t.Run("slide", func(t *testing.T) {
		var (
			buf []byte
			lrd *Reader
		)

		t.Parallel()

		buf = append(
			bytes.Repeat([]byte{'A'}, readSize*2),
			bytes.Repeat([]byte{'B'}, 10)...,
		)

		lrd = NewReader(bytes.NewReader(buf))
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, readSize, lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf[:lrd.head], lrd.buf)

		lrd.current = lrd.head
		lrd.fill()

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, readSize*2, lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf[:lrd.head], lrd.buf)

		lrd.start = lrd.head
		lrd.current = lrd.head
		lrd.fill()
		slices.Reverse(buf)

		assert.Equal(t, nil, lrd.Err())
		assert.Equal(t, 0, lrd.start)
		assert.Equal(t, 0, lrd.current)
		assert.Equal(t, 10, lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf[:initBufSize], lrd.buf)

		lrd.current = lrd.head
		lrd.fill()

		assert.Equal(t, io.EOF, lrd.Err())
		assert.Equal(t, 0, lrd.start)
		assert.Equal(t, 10, lrd.current)
		assert.Equal(t, 10, lrd.head)
		assert.Equal(t, initBufSize, len(lrd.buf))
		assertBuf(t, buf[:initBufSize], lrd.buf)
	})

	t.Run("bogusReader", func(t *testing.T) {
		t.Parallel()

		assert.PanicsWithValue(
			t,
			"langengine/lexer: bogus io.Reader",
			func() {
				NewReader(bogusReader{}).fill()
			},
		)
	})
}

func TestReaderNext(t *testing.T) {
	type testData struct {
		content string
		history []snapshot
	}

	var (
		testTbl []testData
		test    testData
	)

	t.Parallel()

	testTbl = []testData{
		{
			content: "abc",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 1},
				{Position{1, 3}, 2},
			},
		},
		{
			content: "qwertyuiop",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 1},
				{Position{1, 3}, 2},
				{Position{1, 4}, 3},
				{Position{1, 5}, 4},
				{Position{1, 6}, 5},
				{Position{1, 7}, 6},
				{Position{1, 8}, 7},
				{Position{1, 9}, 8},
				{Position{1, 10}, 9},
			},
		},
		{
			// 😀 U+1F600 GRINNING FACE (4 bytes)
			content: "😀😀abc😀😀\n😀",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 4},
				{Position{1, 3}, 8},
				{Position{1, 4}, 9},
				{Position{1, 5}, 10},
				{Position{1, 6}, 11},
				{Position{1, 7}, 15},
				{Position{1, 8}, 19},
				{Position{2, 1}, 20},
			},
		},
		{
			// 中 U+4E2D (3 bytes)
			// 文 U+6587 (3 bytes)
			content: "中文a",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 3},
				{Position{1, 3}, 6},
			},
		},
		{
			// 🐍 U+1F40D (4 bytes)
			content: "go🐍lang",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 1},
				{Position{1, 3}, 2},
				{Position{1, 4}, 6},
				{Position{1, 5}, 7},
				{Position{1, 6}, 8},
				{Position{1, 7}, 9},
				{Position{1, 8}, 9},
			},
		},
		{
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			// 😀 U+1F600 (4 bytes)
			content: "Aé中😀B",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 1},
				{Position{1, 3}, 3},
				{Position{1, 4}, 6},
				{Position{1, 5}, 10},
			},
		},
		{
			// 😀 U+1F600 (4 bytes)
			// 文 U+6587 (3 bytes)
			// 🐍 U+1F40D (4 bytes)
			content: "😀\n文\n🐍a",
			history: []snapshot{
				{Position{1, 1}, 0},
				{Position{1, 2}, 4},
				{Position{2, 1}, 5},
				{Position{2, 2}, 8},
				{Position{3, 1}, 9},
				{Position{3, 2}, 13},
			},
		},
	}

	for _, test = range testTbl {
		t.Run(fmt.Sprintf("%q", test.content), func(t *testing.T) {
			var (
				lrd  *Reader
				char rune
				end  int
			)

			lrd = NewReader(strings.NewReader(test.content))
			end = 1

			for _, char = range test.content {
				assert.Equal(t, char, lrd.Next())
				assert.Equal(t, len(test.content), lrd.head)
				assert.Equal(t, initBufSize, len(lrd.buf))
				assertBuf(t, []byte(test.content), lrd.buf)
				assert.Equal(t, test.history[:end], lrd.history)

				end++
			}

			assert.Equal(t, EOF, lrd.Next())
		})
	}
}
