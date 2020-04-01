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

/*
Package parser implements a parser for keeper files. Input may be
provided in a variety of forms (see the various Parse* functions); the
output is an abstract syntax tree (AST). The parser is invoked through
one of the Parse* functions.

The parser accepts a larger language than is syntactically permitted,
for simplicity, and for improved robustness in the presence of syntax
errors.
*/
package parser

import (
	"errors"
	"fmt"

	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/scanner"
	"go.felesatra.moe/keeper/kpr/token"
)

// A Mode value is a set of flags (or 0). They control the amount of
// source code parsed and other optional parser functionality.
type Mode uint

/*
ParseBytes parses the contents of a keeper file and
returns the corresponding ast.Entry nodes.

ParseBytes parses the source from src and the filename
is only used when recording position information.

The mode parameter controls the amount of source text parsed and other
optional parser functionality. Position information is recorded in the
file set fset, which must not be nil.

If syntax errors were found, the result is a partial AST (with
ast.Bad* nodes representing the fragments of erroneous source
code). Multiple errors are returned via a scanner.ErrorList which is
sorted by file position.
*/
func ParseBytes(fset *token.FileSet, filename string, src []byte, mode Mode) ([]ast.Entry, error) {
	p := &parser{
		f: fset.AddFile(filename, -1, len(src)),
	}
	p.s.Init(p.f, src, p.errs.Add, 0)
	entries := p.parse()
	return entries, p.errs.Err()
}

type parser struct {
	f           *token.File
	s           scanner.Scanner
	errs        scanner.ErrorList
	tokenBuffer []tokenInfo
}

type tokenInfo struct {
	pos token.Pos
	tok token.Token
	lit string
}

// Helper methods

// scan calls Scan on the underlying scanner.
func (p *parser) scan() (token.Pos, token.Token, string) {
	if len(p.tokenBuffer) > 0 {
		t := p.tokenBuffer[len(p.tokenBuffer)-1]
		p.tokenBuffer = p.tokenBuffer[:len(p.tokenBuffer)-1]
		return t.pos, t.tok, t.lit
	}
	return p.s.Scan()
}

func (p *parser) unread(pos token.Pos, tok token.Token, lit string) {
	p.tokenBuffer = append(p.tokenBuffer, tokenInfo{pos, tok, lit})
}

func (p *parser) peek() (token.Pos, token.Token, string) {
	pos, tok, lit := p.scan()
	p.unread(pos, tok, lit)
	return pos, tok, lit
}

// scanLine scans up to and including the next newline
// and returns the position of the newline (or EOF) token.
func (p *parser) scanLine() token.Pos {
	for {
		pos, tok, lit := p.scan()
		if tok == token.EOF {
			p.unread(pos, tok, lit)
			return pos
		}
		if tok != token.NEWLINE {
			continue
		}
		return pos
	}
}

// scanUntilEntry scans until before the beginning of the next
// potential entry (or EOF) and returns the position of the preceding
// newline token.
func (p *parser) scanUntilEntry() token.Pos {
	for {
		startPos := p.scanLine()
		switch pos, tok, lit := p.peek(); {
		default:
			continue
		case tok == token.EOF:
			p.unread(pos, tok, lit)
			return startPos
		case isEntryKeyword(tok):
			return startPos
		}
	}
}

func isEntryKeyword(tok token.Token) bool {
	switch tok {
	case token.TX, token.BALANCE, token.UNIT:
		return true
	default:
		return false
	}
}

func (p *parser) scanLineAsBad(from token.Pos) ast.BadLine {
	return ast.BadLine{From: from, To: p.scanLine()}
}

func (p *parser) scanLineAsBadEntry(from token.Pos) ast.BadEntry {
	return ast.BadEntry{From: from, To: p.scanLine()}
}

// scanUntilEntryAsBad scans until the next entry and returns a BadEntry for
// the intervening tokens.
func (p *parser) scanUntilEntryAsBad(from token.Pos) ast.BadEntry {
	return ast.BadEntry{From: from, To: p.scanUntilEntry()}
}

func (p *parser) errorf(pos token.Pos, format string, v ...interface{}) {
	p.errs.Add(p.f.Position(pos), fmt.Sprintf(format, v...))
}

// Parsing methods

func (p *parser) parse() []ast.Entry {
	var entries []ast.Entry
	for {
		switch pos, tok, lit := p.scan(); {
		case tok == token.EOF:
			return entries
		case tok == token.NEWLINE:
		case isEntryKeyword(tok):
			e := p.parseEntry(pos, tok, lit)
			entries = append(entries, e)
		default:
			p.errorf(pos, "bad token %s %s", tok, lit)
			e := p.scanUntilEntryAsBad(pos)
			entries = append(entries, e)
		}
	}
}

func (p *parser) parseEntry(pos token.Pos, tok token.Token, lit string) ast.Entry {
	switch tok {
	case token.TX:
		return p.parseTransaction(pos)
	case token.UNIT:
		return p.parseUnitDecl(pos)
	case token.BALANCE:
		return p.parseBalance(pos)
	default:
		p.errorf(pos, "bad entry keyword %s", lit)
		return p.scanUntilEntryAsBad(pos)
	}
}

func (p *parser) parseTransaction(pos token.Pos) ast.Entry {
	t := ast.Transaction{
		TokPos: pos,
	}

	pos, tok, lit := p.scan()
	if tok != token.DATE {
		p.errorf(pos, "in transaction expected DATE not %s %s", tok, lit)
		return p.scanUntilEntryAsBad(t.Pos())
	}
	t.Date = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	if tok != token.STRING {
		p.errorf(pos, "in transaction expected STRING not %s %s", tok, lit)
		return p.scanUntilEntryAsBad(t.Pos())
	}
	t.Description = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	if tok != token.NEWLINE {
		p.errorf(pos, "in transaction expected NEWLINE not %s %s", tok, lit)
		return p.scanUntilEntryAsBad(t.Pos())
	}

	var err error
	t.Splits, err = p.parseSplits()
	if err != nil {
		// parseSplits already reported the error.
		return p.scanUntilEntryAsBad(t.Pos())
	}

	pos, tok, lit = p.scan()
	if tok != token.END {
		panic("unexpected token")
	}
	t.EndTok = ast.End{TokPos: pos}

	pos, tok, lit = p.scan()
	if tok != token.NEWLINE {
		p.errorf(pos, "after end bad token %s %s", tok, lit)
		_ = p.scanLine()
	}

	return t
}

func (p *parser) parseSplits() ([]ast.LineNode, error) {
	var splits []ast.LineNode
	for {
		switch pos, tok, lit := p.scan(); tok {
		case token.ACCOUNT:
			p.unread(pos, tok, lit)
			s := p.parseSplit()
			splits = append(splits, s)
		case token.NEWLINE:
			continue
		case token.END:
			p.unread(pos, tok, lit)
			return splits, nil
		case token.EOF:
			p.unread(pos, tok, lit)
			p.errorf(pos, "EOF in transaction")
			return nil, errors.New("EOF in transaction")
		default:
			p.errorf(pos, "in split bad token %s %s", tok, lit)
			n := p.scanLineAsBad(pos)
			splits = append(splits, n)
		}
	}
}

func (p *parser) parseSplit() ast.LineNode {
	var s ast.Split
	pos, tok, lit := p.scan()
	if tok != token.ACCOUNT {
		panic(fmt.Sprintf("unexpected %s %s in split", tok, lit))
	}
	s.Account = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	switch tok {
	case token.NEWLINE:
		return s
	case token.DECIMAL:
	default:
		p.errorf(pos, "in split bad token %s %s", tok, lit)
		return p.scanLineAsBad(s.Pos())
	}
	p.unread(pos, tok, lit)
	a, err := p.parseAmount()
	if err != nil {
		return p.scanLineAsBad(s.Pos())
	}
	s.Amount = &a
	pos, tok, lit = p.scan()
	if tok != token.NEWLINE {
		p.errorf(pos, "in split bad token %s %s", tok, lit)
		return p.scanLineAsBad(s.Pos())
	}
	return s
}

// parseAmount parses an amount.  If parsing fails, the scanning state
// is returned to just before the offending token.
// The error is reported via errorf.
func (p *parser) parseAmount() (ast.Amount, error) {
	var a ast.Amount
	pos, tok, lit := p.scan()
	if tok != token.DECIMAL {
		p.unread(pos, tok, lit)
		p.errorf(pos, "in amount expected DECIMAL not %s %s", tok, lit)
		return a, fmt.Errorf("bad token %s", tok)
	}
	a.Decimal = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	if tok != token.UNIT_SYM {
		p.unread(pos, tok, lit)
		p.errorf(pos, "in amount expected UNIT_SYM not %s %s", tok, lit)
		return a, fmt.Errorf("bad token %s", tok)
	}
	a.Unit = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}
	return a, nil
}

func (p *parser) parseUnitDecl(pos token.Pos) ast.Entry {
	u := ast.UnitDecl{
		TokPos: pos,
	}

	pos, tok, lit := p.scan()
	if tok != token.UNIT_SYM {
		p.errorf(pos, "in unit decl expected UNIT_SYM not %s %s", tok, lit)
		p.unread(pos, tok, lit)
		return p.scanLineAsBadEntry(u.Pos())
	}
	u.Unit = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	if tok != token.DECIMAL {
		p.errorf(pos, "in unit decl expected DECIMAL not %s %s", tok, lit)
		p.unread(pos, tok, lit)
		return p.scanLineAsBadEntry(u.Pos())
	}
	u.Scale = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	if tok != token.NEWLINE {
		p.errorf(pos, "in unit decl expected NEWLINE not %s %s", tok, lit)
		return p.scanLineAsBadEntry(u.Pos())
	}
	return u
}

func (p *parser) parseBalance(pos token.Pos) ast.Entry {
	h := ast.BalanceHeader{
		TokPos: pos,
	}

	pos, tok, lit := p.scan()
	if tok != token.DATE {
		p.errorf(pos, "in balance expected DATE not %s %s", tok, lit)
		p.unread(pos, tok, lit)
		return p.scanLineAsBadEntry(h.Pos())
	}
	h.Date = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	if tok != token.ACCOUNT {
		p.errorf(pos, "in balance expected ACCOUNT not %s %s", tok, lit)
		p.unread(pos, tok, lit)
		return p.scanLineAsBadEntry(h.Pos())
	}
	h.Account = ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}

	pos, tok, lit = p.scan()
	switch tok {
	case token.DECIMAL:
		p.unread(pos, tok, lit)
		return p.parseBalanceSingleAmount(h)
	case token.NEWLINE:
		return p.parseBalanceMultipleAmounts(h)
	default:
		p.errorf(pos, "in balance unexpected %s %s", tok, lit)
		p.unread(pos, tok, lit)
		return p.scanUntilEntryAsBad(h.Pos())
	}
}

// parseBalanceSingleAmount parses the remainder of a single amount balance.
// If parsing fails, the scanning state is returned to just before the
// offending token.
func (p *parser) parseBalanceSingleAmount(h ast.BalanceHeader) ast.Entry {
	b := ast.SingleBalance{
		BalanceHeader: h,
	}
	a, err := p.parseAmount()
	if err != nil {
		return p.scanLineAsBadEntry(b.Pos())
	}
	b.Amount = a

	pos, tok, lit := p.scan()
	if tok != token.NEWLINE {
		p.errorf(pos, "in balance expected NEWLINE not %s %s", tok, lit)
		return p.scanLineAsBadEntry(b.Pos())
	}
	return b
}

func (p *parser) parseBalanceMultipleAmounts(h ast.BalanceHeader) ast.Entry {
	b := ast.MultiBalance{
		BalanceHeader: h,
	}
	for {
		switch pos, tok, lit := p.scan(); tok {
		case token.DECIMAL:
			p.unread(pos, tok, lit)
			a, err := p.parseAmount()
			if err != nil {
				a := p.scanLineAsBad(pos)
				b.Amounts = append(b.Amounts, a)
				continue
			}
			b.Amounts = append(b.Amounts, ast.AmountLine{Amount: a})
		case token.NEWLINE:
			continue
		case token.END:
			b.EndTok = ast.End{TokPos: pos}
			pos, tok, lit = p.scan()
			if tok != token.NEWLINE {
				p.errorf(pos, "after end bad token %s %s", tok, lit)
				_ = p.scanLine()
			}
			return b
		case token.EOF:
			p.unread(pos, tok, lit)
			p.errorf(pos, "EOF in multi-line balance")
			return p.scanUntilEntryAsBad(h.Pos())
		default:
			p.errorf(pos, "in balance bad token %s %s", tok, lit)
			a := p.scanLineAsBad(pos)
			b.Amounts = append(b.Amounts, a)
		}
	}
}
