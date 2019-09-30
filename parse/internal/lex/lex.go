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

package lex

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Lexer struct {
	r       *bufio.Reader
	lastPos lexPos
	pos     lexPos
	pending []rune
	state   stateFn
	tokens  chan Token
}

func Lex(r io.Reader) *Lexer {
	return &Lexer{
		r:       bufio.NewReader(r),
		lastPos: lexPos{line: 1},
		pos:     lexPos{line: 1},
		state:   lexStart,
		tokens:  make(chan Token, 2),
	}
}

func (l *Lexer) NextToken() (t Token) {
	defer l.recover(&t)
	for {
		select {
		case tok := <-l.tokens:
			return tok
		default:
			if l.state == nil {
				return Token{Typ: TokEOF}
			}
			l.state = l.state(l)
		}
	}
}

// next reads and returns the next rune.
func (l *Lexer) next() rune {
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

// unread unreads the last rune returned by next.
func (l *Lexer) unread() {
	if l.pos == l.lastPos {
		panic(l.pos)
	}
	_ = l.r.UnreadRune()
	l.pending = l.pending[:len(l.pending)-1]
	l.pos = l.lastPos
}

func (l *Lexer) peek() rune {
	r, _, err := l.r.ReadRune()
	if err != nil {
		panic(readErr{err: err})
	}
	_ = l.r.UnreadRune()
	return r
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.unread()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.unread()
}

// emit pending runes as a token.
func (l *Lexer) emit(typ TokenType) {
	l.tokens <- Token{
		Typ: typ,
		Val: string(l.pending),
	}
	l.pending = l.pending[:0]
}

// ignore all pending runes.
func (l *Lexer) ignore() {
	l.pending = l.pending[:0]
}

// errorf emits an error token and returns an exit stateFn.
func (l *Lexer) errorf(format string, v ...interface{}) stateFn {
	l.tokens <- Token{
		Typ: TokError,
		Val: fmt.Sprintf(format, v...),
	}
	return nil
}

// recover recovers from readErr panics.
// This simplifies internal error handling.
func (l *Lexer) recover(t *Token) {
	v := recover()
	if v == nil {
		return
	}
	if v, ok := v.(readErr); ok {
		l.state = nil
		if v.err == io.EOF {
			*t = Token{Typ: TokEOF}
		} else {
			*t = Token{
				Typ: TokError,
				Val: fmt.Sprintf("Read error: %v", v.err),
			}
		}
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

type stateFn func(*Lexer) stateFn

// readErr is passed to panic to signal read errors.
// This is caught by lexer.recover.
type readErr struct {
	err error
}

func lexStart(l *Lexer) stateFn {
	switch r := l.next(); {
	case r == '.':
		l.emit(TokDot)
		return lexStart
	case r == '\n':
		l.emit(TokNewline)
		return lexStart
	case r == '"':
		return lexString
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

func lexDecimal(l *Lexer) stateFn {
	l.acceptRun(digits)
	l.accept(".")
	return lexDecimalAfterPoint
}

func lexDecimalAfterPoint(l *Lexer) stateFn {
	l.acceptRun(digits)
	if r := l.peek(); !unicode.IsSpace(r) {
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
	l.emit(TokDecimal)
	return lexStart
}

func lexDigit(l *Lexer) stateFn {
	l.acceptRun(digits)
	switch r := l.next(); {
	case r == '.':
		return lexDecimalAfterPoint
	case r == '-':
		return lexDate
	case unicode.IsSpace(r):
		l.unread()
		l.emit(TokDecimal)
		return lexStart
	default:
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
}

func lexDate(l *Lexer) stateFn {
	l.acceptRun(digits + "-")
	if r := l.peek(); !unicode.IsSpace(r) {
		return l.errorf("unexpected %v at %v", r, l.pos)
	}
	l.emit(TokDate)
	return lexStart
}

func lexUpper(l *Lexer) stateFn {
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
			l.emit(TokUnit)
			return lexStart
		default:
			return l.errorf("unexpected char %v at %v", r, l.pos)
		}
	}
}

func lexLower(l *Lexer) stateFn {
	l.acceptRun(letters)
	switch r := l.next(); {
	case r == ':':
		return lexAccount
	case unicode.IsSpace(r):
		l.unread()
		l.emit(TokKeyword)
		return lexStart
	default:
		return l.errorf("unexpected char %v at %v", r, l.pos)
	}
}

func lexAccount(l *Lexer) stateFn {
	l.acceptRun(letters + ":")
	if r := l.peek(); !unicode.IsSpace(r) {
		return l.errorf("unexpected %v at %v", r, l.pos)
	}
	l.emit(TokAccount)
	return lexStart
}

func lexString(l *Lexer) stateFn {
	for {
		switch r := l.next(); r {
		case '"':
			l.emit(TokString)
			return lexStart
		case '\\':
			_ = l.next()
		case '\n':
			return l.errorf("unclosed string at %v", l.pos)
		}
	}
}
