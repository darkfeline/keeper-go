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

// Package scanner implements a scanner for kpr files. It takes
// a []byte as source which can then be tokenized through repeated
// calls to the Scan method.
package scanner

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.felesatra.moe/keeper/kpr/token"
)

// An ErrorHandler may be provided to Scanner.Init. If a syntax error
// is encountered and a handler was installed, the handler is called
// with a position and an error message. The position points to the
// beginning of the offending token.
type ErrorHandler func(token.Position, string)

// A Scanner holds the scanner's internal state while processing a
// given text. It can be allocated as part of another data structure
// but must be initialized via Init before use.
type Scanner struct {
	// Static state
	f    *token.File
	src  []byte
	err  ErrorHandler
	mode Mode

	// Scanning state
	start      int // starting offset of pending
	offset     int // current scan offset
	pending    []rune
	state      stateFn
	results    chan result
	scannedEOF bool

	// Public state - ok to modify
	ErrorCount int
}

// A mode value is a set of flags (or 0). They control scanner behavior.
type Mode uint

const (
	ScanComments Mode = 1 << iota // return comments as COMMENT tokens
)

type result struct {
	Pos token.Pos
	Tok token.Token
	Lit string
}

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src. The scanner uses the file set file
// for position information and it adds line information for each
// line. It is ok to re-use the same file when re-scanning the same
// file as line information which is already present is ignored. Init
// causes a panic if the file size does not match the src size.
//
// Calls to Scan will invoke the error handler err if they encounter a
// syntax error and err is not nil. Also, for each error encountered,
// the Scanner field ErrorCount is incremented by one. The mode
// parameter determines how comments are handled.
//
// Note that Init may call err if there is an error in the first
// character of the file.
func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	if file.Size() != len(src) {
		panic("src size does not match file")
	}
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

/*
Scan scans the next token and returns the token position, the token,
and its literal string if applicable. The source end is indicated by
token.EOF.

In all cases, the literal string is the scanned token.

For more tolerant parsing, Scan will return a valid token if possible
even if a syntax error was encountered. Thus, even if the resulting
token sequence contains no illegal tokens, a client may not assume
that no error occurred. Instead it must check the scanner's ErrorCount
or the number of calls of the error handler, if there was one
installed.

Scan adds line information to the file added to the file set with
Init. Token positions are relative to that file and thus relative to
the file set.
*/
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	for {
		select {
		case r := <-s.results:
			return r.Pos, r.Tok, r.Lit
		default:
			if s.state == nil {
				panic("nil state in scanner")
			}
			s.state = s.state(s)
		}
	}
}

const eof rune = -1

// next reads and returns the next rune, which may be invalid.
// utf8.RuneError is returned for decoding errors.
func (s *Scanner) next() rune {
	if s.offset >= len(s.src) {
		s.pending = append(s.pending, eof)
		return eof
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
	case eof:
	case utf8.RuneError:
		s.offset -= 1
	default:
		s.offset -= utf8.RuneLen(r)
	}
	s.pending = s.pending[:last]
}

func (s *Scanner) peek() rune {
	r := s.next()
	s.unread()
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
	for s.accept(valid) {
	}
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

func (s *Scanner) emitEOF() {
	s.ignore()
	s.results <- result{
		Pos: s.f.Pos(len(s.src)),
		Tok: token.EOF,
		Lit: "",
	}
}

// ignore throws away all pending runes.
func (s *Scanner) ignore() {
	s.pending = s.pending[:0]
	s.start = s.offset
}

// record an error.
func (s *Scanner) errorf(offset int, format string, v ...interface{}) {
	s.ErrorCount++
	if s.err == nil {
		return
	}
	s.err(s.f.Position(s.f.Pos(offset)), fmt.Sprintf(format, v...))
}

const (
	digits  = "0123456789"
	lower   = "abcdefghijklmnopqrstuvwxyz"
	upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letters = lower + upper
)

type stateFn func(*Scanner) stateFn

func lexStart(s *Scanner) stateFn {
	switch r := s.next(); {
	default:
		s.errorf(s.offset-1, "bad rune %c at start of token", r)
		return lexIllegal
	case r == eof:
		if s.scannedEOF {
			panic("scanned beyond EOF")
		}
		s.scannedEOF = true
		s.emitEOF()
		return lexStart
	case r == '#':
		return lexComment
	case r == '\n':
		s.emit(token.NEWLINE)
		return lexStart
	case r == '"':
		return lexString
	case unicode.IsUpper(r):
		return lexUpper
	case unicode.IsLower(r):
		return lexLower
	case unicode.IsDigit(r):
		return lexDigit
	case r == '-':
		return lexDecimal
	case unicode.IsSpace(r):
		s.ignore()
		return lexStart
	}
}

func lexIllegal(s *Scanner) stateFn {
	s.acceptNonSpace()
	s.emit(token.ILLEGAL)
	return lexStart
}

func lexComment(s *Scanner) stateFn {
	for {
		r := s.next()
		if r != '\n' {
			continue
		}
		s.unread()
		if s.mode&ScanComments != 0 {
			s.emit(token.COMMENT)
		} else {
			s.ignore()
		}
		return lexStart
	}
}

// Expression-like tokens cannot be followed by expression-like characters.
func lexExprEnd(s *Scanner) stateFn {
	switch next := s.peek(); {
	case unicode.IsLetter(next), unicode.IsDigit(next):
		fallthrough
	case next == '-', next == ':', next == '_':
		s.errorf(s.offset, "token followed by non-space %c", next)
	}
	return lexStart
}

func lexString(s *Scanner) stateFn {
	for {
		switch r := s.next(); r {
		case '"':
			s.emit(token.STRING)
			return lexExprEnd
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

func lexUpper(s *Scanner) stateFn {
	for {
		switch r := s.next(); {
		case unicode.IsUpper(r):
		case unicode.IsDigit(r):
			return lexAccountName
		case unicode.IsLower(r):
			return lexAccountName
		case r == '_':
			return lexAccountName
		case r == ':':
			return lexAccountName
		default:
			s.unread()
			s.emit(token.USYMBOL)
			return lexExprEnd
		}
	}
}

func lexLower(s *Scanner) stateFn {
	s.acceptRun(letters)
	switch pending := string(s.pending); pending {
	case "tx":
		s.emit(token.TX)
		return lexExprEnd
	case "end":
		s.emit(token.END)
		return lexExprEnd
	case "balance":
		s.emit(token.BALANCE)
		return lexExprEnd
	case "unit":
		s.emit(token.UNIT)
		return lexExprEnd
	case "disable":
		s.emit(token.DISABLE)
		return lexExprEnd
	}
	if s.accept(digits + ":_") {
		return lexAccountName
	}
	s.errorf(s.start, "invalid token")
	s.emit(token.ILLEGAL)
	return lexExprEnd
}

func lexLetter(s *Scanner) stateFn {
	s.acceptRun(letters)
	switch pending := string(s.pending); {
	case pending == "tx":
		s.emit(token.TX)
		return lexExprEnd
	case pending == "balance":
		s.emit(token.BALANCE)
		return lexExprEnd
	case pending == "unit":
		s.emit(token.UNIT)
		return lexExprEnd
	case s.accept(digits + ":_"):
		return lexAccountName
	case isUpper(pending):
		s.emit(token.USYMBOL)
		return lexExprEnd
	default:
		s.errorf(s.start, "invalid token")
		s.emit(token.ILLEGAL)
		return lexExprEnd
	}
}

func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func lexAccountName(s *Scanner) stateFn {
	s.acceptRun(letters + digits + ":_")
	if pending := string(s.pending); !strings.Contains(pending, ":") {
		s.errorf(s.start, "invalid token")
		s.emit(token.ILLEGAL)
		return lexExprEnd
	}
	s.emit(token.ACCTNAME)
	return lexExprEnd
}

func lexDigit(s *Scanner) stateFn {
	s.acceptRun(digits)
	switch r := s.next(); {
	case r == ',', r == '.':
		return lexDecimal
	case r == '-':
		return lexDate
	default:
		s.unread()
		s.emit(token.DECIMAL)
		return lexExprEnd
	}
}

func lexDecimal(s *Scanner) stateFn {
	s.acceptRun(digits + ".,")
	s.emit(token.DECIMAL)
	return lexExprEnd
}

func lexDate(s *Scanner) stateFn {
	s.acceptRun(digits + "-")
	s.emit(token.DATE)
	return lexExprEnd
}
