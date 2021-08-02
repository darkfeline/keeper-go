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

// Package parser implements a parser for kpr files. Input may be
// provided in a variety of forms (see the various Parse* functions);
// the output is an abstract syntax tree (AST). The parser is invoked
// through one of the Parse* functions.
package parser

import (
	"fmt"
	"strings"

	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/scanner"
	"go.felesatra.moe/keeper/kpr/token"
)

// A Mode value is a set of flags (or 0). They control the amount of
// source code parsed and other optional parser functionality.
type Mode uint

const (
	ParseComments Mode = 1 << iota
)

// ParseBytes parses the contents of a keeper file and returns the
// corresponding ast.Entry nodes.
//
// ParseBytes parses the source from src and the filename is only used
// when recording position information.
//
// The mode parameter controls the amount of source text parsed and
// other optional parser functionality. Position information is
// recorded in the file set fset, which must not be nil.
//
// If syntax errors were found, the result is a partial AST (with
// ast.Bad* nodes representing the fragments of erroneous source
// code). Multiple errors are returned via a scanner.ErrorList which
// is sorted by file position.
func ParseBytes(fset *token.FileSet, filename string, src []byte, mode Mode) (*ast.File, error) {
	p := &parser{
		f:             fset.AddFile(filename, -1, len(src)),
		parseComments: mode&ParseComments != 0,
	}
	var m scanner.Mode
	if p.parseComments {
		m |= scanner.ScanComments
	}
	p.lp = newLineParser(p.f, src, p.errs.Add, m)
	p.parse()
	return &ast.File{
		Entries:  p.entries,
		Comments: p.comments,
	}, p.errs.Err()
}

type parser struct {
	f             *token.File
	parseComments bool
	lp            *lineParser
	current       *line
	curGroup      *ast.CommentGroup

	entries  []ast.Entry
	comments []*ast.CommentGroup
	errs     scanner.ErrorList
}

func (p *parser) nextLine() *line {
	l := p.lp.parseLine()
	if p.parseComments {
		if l.comment == nil || len(l.tokens) > 0 {
			p.curGroup = nil
		}
		if l.comment != nil {
			if p.curGroup == nil {
				p.curGroup = &ast.CommentGroup{}
				p.comments = append(p.comments, p.curGroup)
			}
			c := p.curGroup
			c.List = append(c.List, l.comment)
		}
	}
	p.current = l
	return l
}

func (p *parser) errorf(pos token.Pos, format string, v ...interface{}) {
	p.errs.Add(p.f.Position(pos), fmt.Sprintf(format, v...))
}

func (p *parser) parse() {
	for {
		if p.current != nil && p.current.EOF() {
			return
		}
		l := p.nextLine()
		if l.Empty() {
			continue
		}
		e := p.parseEntry(l)
		p.entries = append(p.entries, e)
	}
}

func (p *parser) parseEntry(l *line) ast.Entry {
	switch l.tokens[0].tok {
	case token.UNIT:
		return p.parseUnitDecl(l)
	case token.TX:
		return p.parseTransaction(l)
	case token.BALANCE, token.TREEBAL:
		return p.parseBalance(l)
	case token.DISABLE:
		return p.parseDisableAccount(l)
	case token.ACCOUNT:
		return p.parseDeclareAccount(l)
	default:
		p.errorf(l.Pos(), "bad entry starting with %s", l.tokens[0].lit)
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
}

func (p *parser) parseUnitDecl(l *line) ast.Entry {
	if err := matchTokens(l.tokens, token.UNIT, token.USYMBOL, token.DECIMAL); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
	u := &ast.UnitDecl{
		TokPos: l.Pos(),
		Unit:   tokVal(l.tokens[1]),
		Scale:  tokVal(l.tokens[2]),
	}
	return u
}

func (p *parser) parseTransaction(l *line) ast.Entry {
	if err := matchTokens(l.tokens, token.TX, token.DATE, token.STRING); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
	e := &ast.Transaction{
		TokPos:      l.Pos(),
		Date:        tokVal(l.tokens[1]),
		Description: tokVal(l.tokens[2]),
	}
	for {
		if p.current.EOF() {
			p.errorf(l.End(), "unexpected EOF")
			return e
		}
		l := p.nextLine()
		if l.Empty() {
			continue
		}
		if err := matchTokens(l.tokens, token.END); err == nil {
			e.EndTok = &ast.End{TokPos: l.Pos()}
			return e
		}
		e.Splits = append(e.Splits, p.parseSplit(l))
	}
}

func (p *parser) parseSplit(l *line) ast.LineNode {
	if err := matchTokens(l.tokens[:1], token.ACCTNAME); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadLine{From: l.Pos(), To: l.End()}
	}
	s := &ast.SplitLine{
		Account: tokVal(l.tokens[0]),
	}
	if len(l.tokens) == 1 {
		return s
	}
	if err := matchTokens(l.tokens, token.ACCTNAME, token.DECIMAL, token.USYMBOL); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadLine{From: l.Pos(), To: l.End()}
	}
	s.Amount = tokAmount(l.tokens[1:])
	return s
}

func (p *parser) parseBalance(l *line) ast.Entry {
	if len(l.tokens) < 3 {
		p.errorf(l.Pos(), "invalid tokens for balance")
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
	if err := matchTokens(l.tokens[1:3], token.DATE, token.ACCTNAME); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
	h := ast.BalanceHeader{
		TokPos:  l.tokens[0].pos,
		Token:   l.tokens[0].tok,
		Date:    tokVal(l.tokens[1]),
		Account: tokVal(l.tokens[2]),
	}

	if err := matchTokens(l.tokens[3:], token.DECIMAL, token.USYMBOL); err == nil {
		return &ast.SingleBalance{
			BalanceHeader: h,
			Amount:        tokAmount(l.tokens[3:]),
		}
	}

	e := &ast.MultiBalance{
		BalanceHeader: h,
	}
	for {
		if p.current.EOF() {
			p.errorf(l.End(), "unexpected EOF")
			return e
		}
		l := p.nextLine()
		if l.Empty() {
			continue
		}
		if err := matchTokens(l.tokens, token.END); err == nil {
			e.EndTok = &ast.End{TokPos: l.Pos()}
			return e
		}
		if err := matchTokens(l.tokens, token.DECIMAL, token.USYMBOL); err != nil {
			p.errorf(l.Pos(), "%s", err)
			n := &ast.BadLine{From: l.Pos(), To: l.End()}
			e.Amounts = append(e.Amounts, n)
			continue
		}
		n := &ast.AmountLine{Amount: tokAmount(l.tokens)}
		e.Amounts = append(e.Amounts, n)
	}
}

func (p *parser) parseDisableAccount(l *line) ast.Entry {
	if err := matchTokens(l.tokens, token.DISABLE, token.DATE, token.ACCTNAME); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
	e := &ast.DisableAccount{
		TokPos:  l.Pos(),
		Date:    tokVal(l.tokens[1]),
		Account: tokVal(l.tokens[2]),
	}
	return e
}

func (p *parser) parseDeclareAccount(l *line) ast.Entry {
	if err := matchTokens(l.tokens, token.ACCOUNT, token.ACCTNAME); err != nil {
		p.errorf(l.Pos(), "%s", err)
		return &ast.BadEntry{From: l.Pos(), To: l.End()}
	}
	e := &ast.DeclareAccount{
		TokPos:  l.Pos(),
		Account: tokVal(l.tokens[1]),
	}
	for {
		if p.current.EOF() {
			p.errorf(l.End(), "unexpected EOF")
			return e
		}
		l := p.nextLine()
		if l.Empty() {
			continue
		}
		if err := matchTokens(l.tokens, token.END); err == nil {
			e.EndTok = &ast.End{TokPos: l.Pos()}
			return e
		}
		if err := matchTokens(l.tokens, token.STRING, token.STRING); err != nil {
			p.errorf(l.Pos(), "%s", err)
			n := &ast.BadLine{From: l.Pos(), To: l.End()}
			e.Metadata = append(e.Metadata, n)
			continue
		}
		n := &ast.MetadataLine{
			Key: tokVal(l.tokens[0]),
			Val: tokVal(l.tokens[1]),
		}
		e.Metadata = append(e.Metadata, n)
	}
}

// Input should start with DECIMAL USYMBOL tokens.
// This function doesn't check the input.
func tokAmount(t []tokenInfo) *ast.Amount {
	if len(t) < 2 || matchTokens(t[:2], token.DECIMAL, token.USYMBOL) != nil {
		panic(fmt.Sprintf("bad tokens %v", t))
	}
	return &ast.Amount{
		Decimal: tokVal(t[0]),
		Unit:    tokVal(t[1]),
	}
}

func tokVal(t tokenInfo) *ast.BasicValue {
	return &ast.BasicValue{ValuePos: t.pos, Kind: t.tok, Value: t.lit}
}

func matchTokens(t []tokenInfo, spec ...token.Token) error {
	if len(t) != len(spec) {
		return &matchError{t, spec}
	}
	for i := 0; i < len(spec); i++ {
		if t[i].tok != spec[i] {
			return &matchError{t, spec}
		}
	}
	return nil
}

// error type to delay string formatting
type matchError struct {
	t    []tokenInfo
	spec []token.Token
}

func (e *matchError) Error() string {
	return fmt.Sprintf("tokens %s do not match %s", formatTokens(e.t), e.spec)
}

func formatTokens(ti []tokenInfo) string {
	var b strings.Builder
	for i, t := range ti {
		if i != 0 {
			b.WriteString(" ")
		}
		b.WriteString(t.tok.String())
	}
	return b.String()
}
