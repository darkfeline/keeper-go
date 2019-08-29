// Copyright (C) 2019  Allen Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stage2

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// tx 2001-02-03 "Some description"
// some:account 123.45 USD
// some:account -123.45 USD
// .
// bal 2001-02-03E4 some:account 123.45 USD

type lexer struct {
	r       *bufio.Reader
	lastPos lexPos
	pos     lexPos
	pending []rune
	state   stateFn
	tokens  chan token
}

func lex(r io.Reader) *lexer {
	return &lexer{
		r:       bufio.NewReader(r),
		lastPos: lexPos{line: 1},
		pos:     lexPos{line: 1},
		state:   lexStart,
		tokens:  make(chan token, 2),
	}
}

func (l *lexer) nextToken() token {
	defer l.recover()
	for {
		select {
		case tok := <-l.tokens:
			return tok
		default:
			if l.state == nil {
				return token{typ: tokEOF}
			}
			l.state = l.state(l)
		}
	}
}

func (l *lexer) next() rune {
	r, _, err := l.r.ReadRune()
	if err != nil {
		panic(readErr{err: err})
	}
	l.pending = append(l.pending, r)
	l.lastPos = l.pos
	if r == '\n' {
		l.pos.line++
		l.pos.col = 0
	} else {
		l.pos.col++
	}
	return r
}

func (l *lexer) unread() {
	if l.pos == l.lastPos {
		panic(l.pos)
	}
	_ = l.r.UnreadRune()
	l.pending = l.pending[:len(l.pending)-1]
	l.pos = l.lastPos
}

func (l *lexer) peek() rune {
	r, _, err := l.r.ReadRune()
	if err != nil {
		panic(readErr{err: err})
	}
	_ = l.r.UnreadRune()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.unread()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.unread()
}

// emit pending runes as a token.
func (l *lexer) emit(typ tokenType) {
	l.tokens <- token{
		typ: typ,
		val: string(l.pending),
	}
	l.pending = l.pending[:0]
}

// ignore all pending runes.
func (l *lexer) ignore() {
	l.pending = l.pending[:0]
}

// errorf emits an error token and returns an exit stateFn.
func (l *lexer) errorf(format string, v ...interface{}) stateFn {
	l.tokens <- token{
		typ: tokError,
		val: fmt.Sprintf(format, v...),
	}
	return nil
}

// recover recovers from readErr panics.
// This simplifies internal error handling.
func (l *lexer) recover() {
	v := recover()
	if v == nil {
		return
	}
	if v, ok := v.(readErr); ok {
		l.state = l.errorf("%v: %v", l.pos, v.err)
		return
	}
	panic(v)
}

type lexPos struct {
	line int
	col  int
}

func (p lexPos) String() string {
	return fmt.Sprintf("line %v col %v", p.line, p.col)
}

type stateFn func(*lexer) stateFn

// readErr is passed to panic to signal read errors.
// This is caught by lexer.recover.
type readErr struct {
	err error
}

type token struct {
	typ tokenType
	val string
}

// go:generate stringer -type=tokenType

type tokenType uint8

const (
	tokError tokenType = iota
	tokEOF
	tokKeyword

	tokOrdering
	tokAccount
	tokString
	tokDecimal
	tokUnit
	tokDot
)

func lexStart(l *lexer) stateFn {
	switch r := l.next(); {
	case r == '"':
		return lexString
	case r == '.':
		l.emit(tokDot)
		return lexStart
	case r == '-':
		return lexDecimal
	case unicode.IsSpace(r):
		l.ignore()
		return lexStart
	case unicode.IsUpper(r):
		return lexUpper
	case unicode.IsLower(r):
		return lexLower
	case unicode.IsDigit(r):
		return lexDigit
	default:
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
}

const (
	digits  = "0123456789"
	lower   = "abcdefghijklmnopqrstuvwxyz"
	upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letters = lower + upper
)

func lexDecimal(l *lexer) stateFn {
	l.acceptRun(digits)
	l.accept(".")
	return lexDecimalAfterPoint
}

func lexDecimalAfterPoint(l *lexer) stateFn {
	l.acceptRun(digits)
	if r := l.peek(); !unicode.IsSpace(r) {
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
	l.emit(tokDecimal)
	return lexStart
}

func lexDigit(l *lexer) stateFn {
	l.acceptRun(digits)
	switch r := l.next(); {
	case r == '.':
		return lexDecimalAfterPoint
	case r == '-':
		return lexOrdering
	case unicode.IsSpace(r):
		l.unread()
		l.emit(tokDecimal)
		return lexStart
	default:
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
}

func lexOrdering(l *lexer) stateFn {
	l.acceptRun(digits + "-E")
	if r := l.peek(); !unicode.IsSpace(r) {
		return l.errorf("unexpected %v at %v", r, l.pos)
	}
	l.emit(tokOrdering)
	return lexStart
}

func lexUpper(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == ':':
			return lexAccount
		case unicode.IsUpper(r):
			continue
		case unicode.IsLower(r):
			return lexLower
		case unicode.IsSpace(r):
			l.unread()
			l.emit(tokUnit)
			return lexStart
		default:
			return l.errorf("unexpected char %v at %v", r, l.pos)
		}
	}
}

func lexLower(l *lexer) stateFn {
	l.acceptRun(letters)
	switch r := l.next(); {
	case r == ':':
		return lexAccount
	case unicode.IsSpace(r):
		l.unread()
		l.emit(tokKeyword)
		return lexStart
	default:
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
}

func lexAccount(l *lexer) stateFn {
	l.acceptRun(letters + ":")
	if r := l.peek(); !unicode.IsSpace(r) {
		return l.errorf("unexpected %v at %v", r, l.pos)
	}
	l.emit(tokAccount)
	return lexStart
}

func lexString(l *lexer) stateFn {
	for {
		switch r := l.next(); r {
		case '"':
			l.emit(tokString)
			return lexStart
		case '\\':
			_ = l.next()
		case '\n':
			return l.errorf("unclosed string at %v", l.pos)
		}
	}
}
