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

package raw

import (
	"fmt"
	"io"

	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/internal/lex"
)

func Parse(r io.Reader) ([]interface{}, error) {
	p := newParser(r)
	return p.parse()
}

type parser struct {
	l *lex.Lexer
}

func newParser(r io.Reader) *parser {
	return &parser{
		l: lex.Lex(r),
	}
}

func (p *parser) parse() ([]interface{}, error) {
	var entries []interface{}
	for {
		switch tok := p.l.NextToken(); tok.Typ {
		case lex.TokEOF:
			return entries, nil
		case lex.TokError:
			return entries, fmt.Errorf("raw: lex error: %v at %v", tok.Val, tok.Pos)
		case lex.TokKeyword:
			v, err := p.parseItem(tok)
			if err != nil {
				return nil, fmt.Errorf("raw: %v", err)
			}
			entries = append(entries, v)
		case lex.TokNewline:
			continue
		default:
			return nil, fmt.Errorf("raw: %v", unexpected(tok))
		}
	}
}

func (p *parser) parseItem(tok lex.Token) (interface{}, error) {
	switch tok.Val {
	case "tx":
		return p.parseTransaction(tok)
	case "unit":
		return p.parseUnit(tok)
	case "bal", "balance":
		return p.parseBalance(tok)
	default:
		return nil, fmt.Errorf("unknown keyword %v at %v", tok.Val, tok.Pos)
	}
}

func (p *parser) parseTransaction(tok lex.Token) (TransactionEntry, error) {
	t := TransactionEntry{Common: Common{Line: tok.Pos.Line}}
	var err error

	tok = p.l.NextToken()
	t.Date, err = parseDateTok(tok)
	if err != nil {
		return t, fmt.Errorf("parse transaction: %v", err)
	}

	tok = p.l.NextToken()
	t.Description, err = parseStringTok(tok)
	if err != nil {
		return t, fmt.Errorf("parse transaction: %v", err)
	}

	tok = p.l.NextToken()
	if tok.Typ != lex.TokNewline {
		return t, fmt.Errorf("parse transaction: %v", unexpected(tok))
	}

	if err := p.parseSplits(&t); err != nil {
		return t, fmt.Errorf("parse transaction: %v", err)
	}
	return t, nil
}

func (p *parser) parseSplits(t *TransactionEntry) error {
	for {
		switch tok := p.l.NextToken(); tok.Typ {
		case lex.TokAccount:
			if err := p.parseSplit(t, tok); err != nil {
				return err
			}
		case lex.TokNewline:
			continue
		case lex.TokDot:
			tok = p.l.NextToken()
			if tok.Typ != lex.TokNewline {
				return unexpected(tok)
			}
			return nil
		default:
			return unexpected(tok)
		}
	}
}
func (p *parser) parseSplit(t *TransactionEntry, tok lex.Token) error {
	var s Split
	var err error
	s.Account = book.Account(tok.Val)

	tok = p.l.NextToken()
	s.Amount, err = p.parseAmount(tok)
	if err != nil {
		return err
	}

	tok = p.l.NextToken()
	if tok.Typ != lex.TokNewline {
		return unexpected(tok)
	}
	t.Splits = append(t.Splits, s)
	return nil
}

func (p *parser) parseUnit(tok lex.Token) (UnitEntry, error) {
	u := UnitEntry{Common: Common{Line: tok.Pos.Line}}
	var err error

	tok = p.l.NextToken()
	u.Symbol, err = parseUnitTok(tok)
	if err != nil {
		return u, fmt.Errorf("parse unit: %v", err)
	}

	tok = p.l.NextToken()
	u.Scale, err = parseDecimalTok(tok)
	if err != nil {
		return u, fmt.Errorf("parse unit: %v", err)
	}

	tok = p.l.NextToken()
	if tok.Typ != lex.TokNewline {
		return u, fmt.Errorf("parse unit: %v", unexpected(tok))
	}
	return u, nil
}

func (p *parser) parseBalance(tok lex.Token) (BalanceEntry, error) {
	b := BalanceEntry{Common: Common{Line: tok.Pos.Line}}
	var err error

	tok = p.l.NextToken()
	b.Date, err = parseDateTok(tok)
	if err != nil {
		return b, fmt.Errorf("parse balance: %v", err)
	}

	tok = p.l.NextToken()
	if tok.Typ != lex.TokAccount {
		return b, fmt.Errorf("parse balance: %v", unexpected(tok))
	}
	b.Account = book.Account(tok.Val)

	tok = p.l.NextToken()
	switch tok.Typ {
	case lex.TokDecimal:
		if err := p.parseBalanceSingleAmount(&b, tok); err != nil {
			return b, fmt.Errorf("parse balance: %v", err)
		}
	case lex.TokNewline:
		if err := p.parseBalanceMultipleAmounts(&b); err != nil {
			return b, fmt.Errorf("parse balance: %v", err)
		}
	default:
		return b, unexpected(tok)
	}
	return b, nil
}

func (p *parser) parseBalanceSingleAmount(b *BalanceEntry, tok lex.Token) error {
	a, err := p.parseAmount(tok)
	if err != nil {
		return err
	}
	b.Amounts = append(b.Amounts, a)

	tok = p.l.NextToken()
	if tok.Typ != lex.TokNewline {
		return unexpected(tok)
	}
	return nil
}

func (p *parser) parseAmount(tok lex.Token) (Amount, error) {
	var a Amount
	var err error
	a.Number, err = parseDecimalTok(tok)
	if err != nil {
		return a, err
	}

	tok = p.l.NextToken()
	a.Unit, err = parseUnitTok(tok)
	if err != nil {
		return a, err
	}
	return a, nil
}

func (p *parser) parseBalanceMultipleAmounts(b *BalanceEntry) error {
	for {
		switch tok := p.l.NextToken(); tok.Typ {
		case lex.TokDecimal:
			if err := p.parseBalanceSingleAmount(b, tok); err != nil {
				return err
			}
		case lex.TokNewline:
			continue
		case lex.TokDot:
			tok = p.l.NextToken()
			if tok.Typ != lex.TokNewline {
				return unexpected(tok)
			}
			return nil
		default:
			return unexpected(tok)
		}
	}
}
