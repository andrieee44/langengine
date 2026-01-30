package lexer_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/andrieee44/langengine/lexer"
	"github.com/stretchr/testify/assert"
)

type helperTestData[T comparable] struct {
	content string
	afterOp string
	op      func(*lexer.Reader) T
	result  T
}

func assertHelperTestDataTbl[T comparable](
	t *testing.T,
	testTbl map[string]helperTestData[T],
) {
	var (
		name string
		test helperTestData[T]
	)

	t.Helper()

	for name, test = range testTbl {
		t.Run(fmt.Sprintf("TestReader%s", name), func(t *testing.T) {
			var (
				lrd    *lexer.Reader
				result T
			)

			lrd = lexer.NewReader(strings.NewReader(test.content))
			result = test.op(lrd)

			assert.Equal(t, test.afterOp, lrd.PeekToken())
			assert.Equal(t, test.result, result)
		})
	}
}

func TestReaderHelpers(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[bool]{
		"Accept": {
			content: "abc",
			afterOp: "a",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("abc")
			},
		},
		"AcceptFunc": {
			content: "Abc",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(unicode.IsLower)
			},
		},
		"AcceptSeq/Empty": {
			content: "",
			afterOp: "",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("")
			},
		},
		"AcceptSeq/ASCII": {
			content: "#define!",
			afterOp: "#define",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("#define")
			},
		},
		"AcceptSeq/Unicode": {
			// 中 U+4E2D (3 bytes)
			// 文 U+6587 (3 bytes)
			content: "中中文b",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("中文")
			},
		},
		"AcceptSeq/Unicode2": {
			// 中 U+4E2D (3 bytes)
			// 文 U+6587 (3 bytes)
			content: "!中中文b",
			afterOp: "!",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				lrd.Next()

				return lrd.AcceptSeq("中文")
			},
		},
	})

	assertHelperTestDataTbl(t, map[string]helperTestData[int]{
		"Until/Greedy": {
			content: "abc,a",
			afterOp: "abc,a",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("")
			},
		},
		"Until/Empty": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("")
			},
		},
		"UntilFunc": {
			content: "abc,",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsPunct)
			},
		},
		"UntilSeq/Empty": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("")
			},
		},
		"UntilSeq/Comment": {
			content: "/* abc */!",
			afterOp: "/* abc ",
			result:  7,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("*/")
			},
		},
		"UntilSeq/Elipsis": {
			content: "..ab..c.d..e...abc",
			afterOp: "..ab..c.d..e",
			result:  12,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("...")
			},
		},
		"UntilSeq/#define": {
			content: "#defin!#definE#ddddd#define",
			afterOp: "#defin!#definE#ddddd",
			result:  20,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("#define")
			},
		},
		"UntilSeq/Unicode": {
			// 😀 U+1F600 (4 bytes)
			// 文 U+6587 (3 bytes)
			// 🐍 U+1F40D (4 bytes)
			content: "文🐍🐍😀文🐍😀",
			afterOp: "文🐍🐍😀",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("文🐍😀")
			},
		},
		"AcceptRun": {
			content: "abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("abc")
			},
		},
		"AcceptRunFunc": {
			content: "ABCa",
			afterOp: "ABC",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsUpper)
			},
		},
	})
}

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

			assert.Equal(t, lexer.EOF, lrd.Next())

			lrd.Backup(test.backups)

			assert.Equal(t, test.afterBackup, lrd.Next())
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

	assert.Equal(t, "ab", token)
	assert.Equal(t, lexer.Position{1, 1}, pos)
	assert.Equal(t, 'c', lrd.Next())

	lrd.Ignore()
	lrd.Next()
	lrd.Next()
	lrd.Next()

	token, pos = lrd.Emit()

	assert.Equal(t, "ABC", token)
	assert.Equal(t, lexer.Position{1, 4}, pos)
	assert.Equal(t, lexer.EOF, lrd.Next())

	lrd = lexer.NewReader(strings.NewReader(""))
	token, pos = lrd.Emit()

	assert.Equal(t, "", token)
	assert.Equal(t, lexer.Position{1, 1}, pos)
	assert.Equal(t, lexer.EOF, lrd.Next())
}

func TestReaderIgnore(t *testing.T) {
	var lrd *lexer.Reader

	t.Parallel()

	lrd = lexer.NewReader(strings.NewReader("abc"))
	lrd.Next()
	lrd.Ignore()

	assert.Equal(t, 'b', lrd.Next())
}
