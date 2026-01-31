package lexer_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/andrieee44/langengine/lexer"
	"github.com/stretchr/testify/assert"
)

type inclusiveResult struct {
	count   int
	matched bool
}

type helperTestData[T comparable] struct {
	content string
	afterOp string
	op      func(*lexer.Reader) T
	result  T
}

func mkInclusiveResult(count int, matched bool) inclusiveResult {
	return inclusiveResult{
		count:   count,
		matched: matched,
	}
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
		t.Run(name, func(t *testing.T) {
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

func TestReaderAccept(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[bool]{
		"Base": {
			content: "abc",
			afterOp: "a",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("abc")
			},
		},
		"NoMatch": {
			content: "abc",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("cde")
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("")
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("abc")
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("")
			},
		},
		"Unicode": {
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			// 😀 U+1F600 (4 bytes)
			content: "é中😀",
			afterOp: "é",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.Accept("é中😀")
			},
		},
	})
}

func TestReaderAcceptFunc(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[bool]{
		"Base": {
			content: "abc",
			afterOp: "a",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(unicode.IsLower)
			},
		},
		"NoMatch": {
			content: "abc",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(unicode.IsUpper)
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(func(rune) bool {
					return false
				})
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(unicode.IsLetter)
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(func(rune) bool {
					return false
				})
			},
		},
		"Unicode": {
			// 😀 U+1F600 (4 bytes)
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			content: "😀é中",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptFunc(unicode.IsLetter)
			},
		},
	})
}

func TestReaderAcceptRun(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[int]{
		"Base": {
			content: "abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("abc")
			},
		},
		"NoMatch": {
			content: "abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("cde")
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("abc")
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("")
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("abc")
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("")
			},
		},
		"Unicode": {
			// 😀 U+1F600 (4 bytes)
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			content: "😀é中",
			afterOp: "😀é中",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("😀é中")
			},
		},
		"UnicodePartialContent": {
			// 😀 U+1F600 (4 bytes)
			// é U+00E9 (2 bytes)
			// 中 U+4E2D (3 bytes)
			content: "😀é中!😀é中",
			afterOp: "😀é中",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRun("😀é中")
			},
		},
	})
}

func TestReaderAcceptRunFunc(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[int]{
		"Base": {
			content: "abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsLower)
			},
		},
		"NoMatch": {
			content: "abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsUpper)
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsLower)
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(func(rune) bool {
					return false
				})
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsGraphic)
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(func(rune) bool {
					return false
				})
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요",
			afterOp: "안녕하세요",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsLetter)
			},
		},
		"UnicodePartialContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요!안녕하세요",
			afterOp: "안녕하세요",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.AcceptRunFunc(unicode.IsLetter)
			},
		},
	})
}

func TestReaderAcceptSeq(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[bool]{
		"Base": {
			content: "abc",
			afterOp: "abc",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("abc")
			},
		},
		"NoMatch": {
			content: "abc",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("cde")
			},
		},
		"PartialMatch": {
			content: "abcd!",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("abcde")
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("abc")
			},
		},
		"PartialMatchContent": {
			content: "abcd!abcde",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("abcde")
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("")
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("abc")
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("")
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요",
			afterOp: "안녕하세요",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("안녕하세요")
			},
		},
		"UnicodePartialMatch": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세!",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("안녕하세요")
			},
		},
		"UnicodePartialContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요!안녕하세요",
			afterOp: "안녕하세요",
			result:  true,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("안녕하세요")
			},
		},
		"UnicodePartialMatchContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세_!안녕하세요",
			afterOp: "",
			result:  false,
			op: func(lrd *lexer.Reader) bool {
				return lrd.AcceptSeq("안녕하세요")
			},
		},
	})
}

func TestReaderUntil(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[int]{
		"Base": {
			content: "abc !",
			afterOp: "abc ",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("!")
			},
		},
		"NoMatch": {
			content: "abc ",
			afterOp: "abc ",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("!")
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("!")
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("")
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("abc")
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("")
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요",
			afterOp: "안녕하세",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("요")
			},
		},
		"UnicodePartialMatchContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요!안녕하세요",
			afterOp: "안녕하세",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.Until("요")
			},
		},
	})
}

func TestReaderUntilFunc(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[int]{
		"Base": {
			content: "abc !",
			afterOp: "abc ",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsPunct)
			},
		},
		"NoMatch": {
			content: "abc ",
			afterOp: "abc ",
			result:  4,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsPunct)
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsPunct)
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(func(rune) bool {
					return true
				})
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsPunct)
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(func(rune) bool {
					return false
				})
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			content: "!!!안",
			afterOp: "!!!",
			result:  3,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsLetter)
			},
		},
		"UnicodePartialMatchContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요!안녕하세요",
			afterOp: "안녕하세요",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilFunc(unicode.IsPunct)
			},
		},
	})
}

func TestReaderUntilFuncInclusive(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[inclusiveResult]{
		"Base": {
			content: "abc !",
			afterOp: "abc !",
			result:  mkInclusiveResult(5, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(unicode.IsPunct),
				)
			},
		},
		"NoMatch": {
			content: "abc ",
			afterOp: "abc ",
			result:  mkInclusiveResult(4, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(unicode.IsPunct),
				)
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc!",
			result:  mkInclusiveResult(4, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(unicode.IsPunct),
				)
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "abc",
			result:  mkInclusiveResult(3, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(func(rune) bool {
						return false
					}),
				)
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  mkInclusiveResult(0, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(unicode.IsPunct),
				)
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  mkInclusiveResult(0, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(func(rune) bool {
						return false
					}),
				)
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			content: "!!!안",
			afterOp: "!!!안",
			result:  mkInclusiveResult(4, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(unicode.IsLetter),
				)
			},
		},
		"UnicodePartialMatchContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "!!!안녕하세요",
			afterOp: "!!!안",
			result:  mkInclusiveResult(4, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(
					lrd.UntilFuncInclusive(unicode.IsLetter),
				)
			},
		},
	})
}

func TestReaderUntilSeq(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[int]{
		"Base": {
			content: "/* abc */",
			afterOp: "/* abc ",
			result:  7,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("*/")
			},
		},
		"NoMatch": {
			content: "/* abc ",
			afterOp: "/* abc ",
			result:  7,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("*/")
			},
		},
		"PartialMatch": {
			content: "abcd!",
			afterOp: "abcd!",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("abcde")
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("abc")
			},
		},
		"PartialMatchContent": {
			content: "abcd!abcde",
			afterOp: "abcd!",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("abcde")
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("")
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("abc")
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("")
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("안녕하세요")
			},
		},
		"UnicodePartialMatch": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세!",
			afterOp: "안녕하세!",
			result:  5,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("안녕하세요")
			},
		},
		"UnicodePartialContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요!안녕하세요",
			afterOp: "",
			result:  0,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("안녕하세요")
			},
		},
		"UnicodePartialMatchContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세_!안녕하세요",
			afterOp: "안녕하세_!",
			result:  6,
			op: func(lrd *lexer.Reader) int {
				return lrd.UntilSeq("안녕하세요")
			},
		},
	})
}

func TestReaderUntilSeqInclusive(t *testing.T) {
	t.Parallel()

	assertHelperTestDataTbl(t, map[string]helperTestData[inclusiveResult]{
		"Base": {
			content: "/* abc */",
			afterOp: "/* abc */",
			result:  mkInclusiveResult(9, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("*/"))
			},
		},
		"NoMatch": {
			content: "/* abc ",
			afterOp: "/* abc ",
			result:  mkInclusiveResult(7, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("*/"))
			},
		},
		"PartialMatch": {
			content: "abcd!",
			afterOp: "abcd!",
			result:  mkInclusiveResult(5, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("abcde"))
			},
		},
		"PartialContent": {
			content: "abc!abc",
			afterOp: "abc",
			result:  mkInclusiveResult(3, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("abc"))
			},
		},
		"PartialMatchContent": {
			content: "abcd!abcde!abcde",
			afterOp: "abcd!abcde",
			result:  mkInclusiveResult(10, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("abcde"))
			},
		},
		"EmptyArgument": {
			content: "abc",
			afterOp: "",
			result:  mkInclusiveResult(0, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive(""))
			},
		},
		"EmptyContent": {
			content: "",
			afterOp: "",
			result:  mkInclusiveResult(0, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("abc"))
			},
		},
		"EmptyAll": {
			content: "",
			afterOp: "",
			result:  mkInclusiveResult(0, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive(""))
			},
		},
		"Unicode": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요",
			afterOp: "안녕하세요",
			result:  mkInclusiveResult(5, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("안녕하세요"))
			},
		},
		"UnicodePartialMatch": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세!",
			afterOp: "안녕하세!",
			result:  mkInclusiveResult(5, false),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("안녕하세요"))
			},
		},
		"UnicodePartialContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세요!안녕하세요",
			afterOp: "안녕하세요",
			result:  mkInclusiveResult(5, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("안녕하세요"))
			},
		},
		"UnicodePartialMatchContent": {
			// 안 U+C548 (3 bytes)
			// 녕 U+B155 (3 bytes)
			// 하 U+D558 (3 bytes)
			// 세 U+C138 (3 bytes)
			// 요 U+D558 (3 bytes)
			content: "안녕하세_!안녕하세요",
			afterOp: "안녕하세_!안녕하세요",
			result:  mkInclusiveResult(11, true),
			op: func(lrd *lexer.Reader) inclusiveResult {
				return mkInclusiveResult(lrd.UntilSeqInclusive("안녕하세요"))
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
	assert.Equal(t, 'b', lrd.Next())
	assert.Equal(t, 'c', lrd.Next())
	assert.Equal(t, lexer.EOF, lrd.Next())
	assert.Equal(t, lexer.EOF, lrd.Peek())
	assert.Equal(t, lexer.EOF, lrd.Peek())
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
