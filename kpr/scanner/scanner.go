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

package scanner

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.felesatra.moe/keeper/kpr/token"
)

type ErrorHandler func(token.Position, string)

type Scanner struct {
	// Static state
	f    *token.File
	src  []byte
	err  ErrorHandler
	mode Mode

	// Scanning state
	start   int // starting offset of pending
	offset  int // current scan offset
	pending []rune
	state   stateFn
	results chan result

	// Public state
	ErrorCount int
}

type Mode uint

const (
	ScanComments Mode = 1 << iota // return comments as COMMENT tokens
)

type result struct {
	Pos token.Pos
	Tok token.Token
	Lit string
}

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	s.f = file
	s.src = src
	s.err = err
	s.mode = mode

	s.start = 0
	s.offset = 0
	s.pending = nil
	s.state = lexStart
	s.results = make(chan result, 2)
	s.ErrorCount = 0
}

func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	for {
		select {
		case r := <-s.results:
			return r.Pos, r.Tok, r.Lit
		default:
			if s.offset >= len(s.src) || s.state == nil {
				return s.f.Pos(s.offset), token.EOF, ""
			}
			s.state = s.state(s)
		}
	}
}

// next reads and returns the next rune, which may be invalid.
// utf8.RuneError is returned for decoding errors.
func (s *Scanner) next() rune {
	if s.offset == len(s.src) {
		panic("next at end of scan buffer")
	}
	r, n := utf8.DecodeRune(s.src[s.offset:])
	s.pending = append(s.pending, r)
	s.offset += n
	if r == '\n' {
		s.f.AddLine(s.offset)
	}
	return r
}

// unread unreads the last rune returned by next.
func (s *Scanner) unread() {
	last := len(s.pending) - 1
	switch r := s.pending[last]; r {
	case utf8.RuneError:
		s.offset -= 1
	default:
		s.offset -= utf8.RuneLen(r)
	}
	s.pending = s.pending[:last]
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.src[s.offset:])
	return r
}

// accept reads the next rune if it is in the valid string.
func (s *Scanner) accept(valid string) bool {
	if strings.IndexRune(valid, s.next()) >= 0 {
		return true
	}
	s.unread()
	return false
}

// acceptRun reads all contiguous runes in the valid string.
func (s *Scanner) acceptRun(valid string) {
	for strings.IndexRune(valid, s.next()) >= 0 {
	}
	s.unread()
}

// acceptNonSpace reads all contiguous runes that are not space.
func (s *Scanner) acceptNonSpace() {
	for unicode.IsSpace(s.next()) {
	}
	s.unread()
}

// emit pending runes as a token.
func (s *Scanner) emit(tok token.Token) {
	s.results <- result{
		Pos: s.f.Pos(s.start),
		Tok: tok,
		Lit: string(s.pending),
	}
	s.ignore()
}

// ignore throws away all pending runes.
func (s *Scanner) ignore() {
	s.pending = s.pending[:0]
	s.start = s.offset
}

func (s *Scanner) errorf(offset int, format string, v ...interface{}) {
	s.ErrorCount++
	if s.err == nil {
		return
	}
	s.err(s.f.Position(s.f.Pos(offset)), fmt.Sprintf(format, v...))
}

type stateFn func(*Scanner) stateFn

func lexStart(s *Scanner) stateFn {
	switch r := s.next(); {
	case r == '#':
		return lexComment
	case r == '\n':
		s.emit(token.NEWLINE)
		return lexStart
	case r == '.':
		s.emit(token.DOT)
		return lexStart
	case r == '"':
		return lexString
	case unicode.IsLetter(r):
		return lexIdent
	case unicode.IsDigit(r):
		return lexDigit
	case r == '-':
		return lexDecimal
	case unicode.IsSpace(r):
		s.ignore()
		return lexStart
	default:
		s.errorf(s.offset-1, "bad rune %c at start of token", r)
		return lexIllegal
	}
}

func lexIllegal(s *Scanner) stateFn {
	s.acceptNonSpace()
	s.emit(token.ILLEGAL)
	return lexStart
}

const (
	digits       = "0123456789"
	lower        = "abcdefghijklmnopqrstuvwxyz"
	upper        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letters      = lower + upper
	identChars   = letters
	accountChars = identChars + digits + ":"
	decimalChars = digits + ","
)

func lexComment(s *Scanner) stateFn {
	for {
		r := s.next()
		if r == '\n' {
			s.unread()
			if s.mode&ScanComments != 0 {
				s.emit(token.COMMENT)
			} else {
				s.ignore()
			}
			s.next()
			s.emit(token.NEWLINE)
			return lexStart
		}
	}
}

func lexString(s *Scanner) stateFn {
	for {
		switch r := s.next(); r {
		case '"':
			s.emit(token.STRING)
			return lexStart
		case '\\':
			s.next()
		case '\n':
			s.unread()
			s.errorf(s.offset, "unclosed string")
			s.emit(token.ILLEGAL)
			return lexStart
		}
	}
}

func lexIdent(s *Scanner) stateFn {
	s.acceptRun(identChars)
	switch r := s.peek(); true {
	case unicode.IsSpace(r):
		s.emit(token.IDENT)
		return lexStart
	case r == ':':
		return lexAccount
	default:
		s.emit(token.IDENT)
		s.errorf(s.offset, "unexpected rune %c in IDENT", r)
		return lexIllegal
	}
}

func lexAccount(s *Scanner) stateFn {
	s.acceptRun(accountChars)
	s.emit(token.ACCOUNT)
	if r := s.peek(); !unicode.IsSpace(r) {
		s.errorf(s.offset, "unexpected rune %c in ACCOUNT", r)
		return lexIllegal
	}
	return lexStart
}

func lexDigit(s *Scanner) stateFn {
	s.acceptRun(digits)
	switch r := s.next(); {
	case r == ',':
		return lexDecimal
	case r == '.':
		return lexDecimalAfterPoint
	case r == '-':
		return lexDate
	case unicode.IsSpace(r):
		s.unread()
		s.emit(token.DECIMAL)
		return lexStart
	default:
		s.unread()
		s.emit(token.DECIMAL)
		s.errorf(s.offset, "unexpected rune %c in DECIMAL", r)
		return lexIllegal
	}
}

func lexDecimal(s *Scanner) stateFn {
	s.acceptRun(decimalChars)
	s.accept(".")
	return lexDecimalAfterPoint
}

func lexDecimalAfterPoint(s *Scanner) stateFn {
	s.acceptRun(decimalChars)
	if r := s.peek(); !unicode.IsSpace(r) {
		s.errorf(s.offset, "unexpected rune %c in DECIMAL", r)
		return lexIllegal
	}
	s.emit(token.DECIMAL)
	return lexStart
}

func lexDate(s *Scanner) stateFn {
	s.acceptRun(digits + "-")
	if r := s.peek(); !unicode.IsSpace(r) {
		s.errorf(s.offset, "unexpected rune %c in DATE", r)
		return lexIllegal
	}
	s.emit(token.DATE)
	return lexStart
}
