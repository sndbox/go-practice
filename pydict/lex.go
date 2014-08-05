// Based on text/template/parse/lex.go

package pydict

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

var _ = fmt.Println

const eof = -1

const (
	itemError = iota
	itemEOF
	itemLeftBrace
	itemRightBrace
	itemLeftBracket
	itemRightBracket
	itemString
	itemComma
	itemColon
)

type itemType int

type item struct {
	iType itemType
	value string
}

type stateFunc func(*lexer) stateFunc

type lexer struct {
	input string
	start int
	width int
	pos   int
	state stateFunc
	out   chan item
}

func lex(input string) *lexer {
	out := make(chan item)
	l := &lexer{input, 0, 0, 0, nil, out}
	go l.run()
	return l
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(i itemType) {
	l.out <- item{i, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) run() {
	for l.state = startState; l.state != nil; {
		l.state = l.state(l)
	}
}

func (l *lexer) nextItem() item {
	i := <-l.out
	return i
}

func startState(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			break
		}
		switch r {
		case '{':
			l.emit(itemLeftBrace)
		case '}':
			l.emit(itemRightBrace)
		case '[':
			l.emit(itemLeftBracket)
		case ']':
			l.emit(itemRightBracket)
		case ',':
			l.emit(itemComma)
		case ':':
			l.emit(itemColon)
		case '\'':
			return stringState
		case '#':
			return commentState
		default:
			l.start += 1
		}
	}
	l.emit(itemEOF)
	return nil
}

// ugly.......
func stringState(l *lexer) stateFunc {
	i := strings.Index(l.input[l.pos:], "'")
	if i < 0 {
		l.emit(itemError)
		l.pos = len(l.input)
	} else {
		l.start += 1
		l.pos += i
		l.emit(itemString)
		l.start += 1
		l.pos += 1
	}
	return startState
}

func commentState(l *lexer) stateFunc {
	i := strings.Index(l.input[l.pos:], "\n")
	if i < 0 {
		l.pos = len(l.input)
	} else {
		l.start = i
		l.pos = i
	}
	return startState
}
