package lexer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/andrieee44/langengine/lexer"
	"github.com/stretchr/testify/assert"
)

func TestReaderBackup(t *testing.T) {
	type testData struct {
		content     string
		afterBackup rune
		backups     int
	}

	var (
		testTbl []testData
		test    testData
	)

	t.Parallel()

	testTbl = []testData{
		{
			content:     "abc",
			afterBackup: 'c',
			backups:     1,
		},
		{
			content:     "abc",
			afterBackup: 'a',
			backups:     999,
		},
		{
			// é U+00E9 (2 bytes)
			content:     "café",
			afterBackup: 'é',
			backups:     1,
		},
		{
			// é U+00E9 (2 bytes)
			content:     "café",
			afterBackup: 'f',
			backups:     2,
		},
		{
			// é U+00E9 (2 bytes)
			content:     "café",
			afterBackup: 'c',
			backups:     999,
		},

		{
			// 中 U+4E2D (3 bytes)
			// 文 U+6587 (3 bytes)
			content:     "中文",
			afterBackup: '文',
			backups:     1,
		},
		{
			// 中 U+4E2D (3 bytes)
			// 文 U+6587 (3 bytes)
			content:     "中文",
			afterBackup: '中',
			backups:     2,
		},
		{
			// 中 U+4E2D (3 bytes)
			// 文 U+6587 (3 bytes)
			content:     "中文",
			afterBackup: '中',
			backups:     999,
		},

		{
			// 😀 U+1F600 (4 bytes)
			content:     "😀go",
			afterBackup: 'o',
			backups:     1,
		},
		{
			// 😀 U+1F600 (4 bytes)
			content:     "😀go",
			afterBackup: 'g',
			backups:     2,
		},
		{
			// 😀 U+1F600 (4 bytes)
			content:     "😀go",
			afterBackup: '😀',
			backups:     999,
		},

		{
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			// 😀 U+1F600 (4 bytes)
			content:     "Aé中😀B",
			afterBackup: 'B',
			backups:     1,
		},
		{
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			// 😀 U+1F600 (4 bytes)
			content:     "Aé中😀B",
			afterBackup: '中',
			backups:     3,
		},
		{
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			// 😀 U+1F600 (4 bytes)
			content:     "Aé中😀B",
			afterBackup: 'A',
			backups:     999,
		},

		{
			// 😀 U+1F600 (4 bytes)
			// 文 U+6587 (3 bytes)
			// 🐍 U+1F40D (4 bytes)
			content:     "😀\n文\n🐍",
			afterBackup: '🐍',
			backups:     1,
		},
		{
			// 😀 U+1F600 (4 bytes)
			// 文 U+6587 (3 bytes)
			// 🐍 U+1F40D (4 bytes)
			content:     "😀\n文\n🐍",
			afterBackup: '文',
			backups:     3,
		},
		{
			// 😀 U+1F600 (4 bytes)
			// 文 U+6587 (3 bytes)
			// 🐍 U+1F40D (4 bytes)
			content:     "😀\n文\n🐍",
			afterBackup: '😀',
			backups:     999,
		},
	}

	for _, test = range testTbl {
		t.Run(fmt.Sprintf("%q", test.content), func(t *testing.T) {
			var lrd *lexer.Reader

			lrd = lexer.NewReader(strings.NewReader(test.content))

			for range test.content {
				lrd.Next()
			}

			assert.Equal(t, lrd.Next(), lexer.EOF)

			lrd.Backup(test.backups)

			assert.Equal(t, lrd.Next(), test.afterBackup)
		})
	}
}

func TestReaderPeek(t *testing.T) {
	var lrd *lexer.Reader

	t.Parallel()

	lrd = lexer.NewReader(strings.NewReader("abc"))

	assert.Equal(t, 'a', lrd.Peek())
	assert.Equal(t, 'a', lrd.Next())
}

func TestReaderEmit(t *testing.T) {
	var (
		lrd   *lexer.Reader
		token string
		pos   lexer.Position
	)

	t.Parallel()

	lrd = lexer.NewReader(strings.NewReader("abcABC"))
	lrd.Next()
	lrd.Next()

	token, pos = lrd.Emit()

	assert.Equal(t, token, "ab")
	assert.Equal(t, pos, lexer.Position{1, 1})
	assert.Equal(t, lrd.Next(), 'c')

	lrd.Ignore()
	lrd.Next()
	lrd.Next()
	lrd.Next()

	token, pos = lrd.Emit()

	assert.Equal(t, token, "ABC")
	assert.Equal(t, pos, lexer.Position{1, 4})
	assert.Equal(t, lrd.Next(), lexer.EOF)

	lrd = lexer.NewReader(strings.NewReader(""))
	token, pos = lrd.Emit()

	assert.Equal(t, token, "")
	assert.Equal(t, pos, lexer.Position{1, 1})
	assert.Equal(t, lrd.Next(), lexer.EOF)
}

func TestReaderIgnore(t *testing.T) {
	var lrd *lexer.Reader

	t.Parallel()

	lrd = lexer.NewReader(strings.NewReader("abc"))
	lrd.Next()
	lrd.Ignore()

	assert.Equal(t, lrd.Next(), 'b')
}
