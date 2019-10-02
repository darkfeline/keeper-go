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

package parse

import (
	"io"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/internal/lex"
	"golang.org/x/xerrors"
)

func Parse(r io.Reader) []book.Transaction {
	return nil
}

type parser struct {
	l     *lex.Lexer
	units map[string]book.UnitType
}

func newParser(r io.Reader) *parser {
	return &parser{
		l:     lex.Lex(r),
		units: make(map[string]book.UnitType),
	}
}

func (p *parser) parse(r io.Reader) ([]interface{}, error) {
	var items []interface{}
	for {
		switch tok := p.l.NextToken(); tok.Typ {
		case lex.TokEOF:
			return items, nil
		case lex.TokError:
			return nil, xerrors.Errorf("keeper/parse: lex error: %v at %v", tok.Val, tok.Pos)
		case lex.TokKeyword:
			v, err := p.parseItem(tok)
			if err != nil {
				return nil, xerrors.Errorf("keeper/parse: %v", err)
			}
			items = append(items, v)
		default:
			return nil, xerrors.Errorf("keeper/parse: unexpected token: %v at %v", tok.Val, tok.Pos)
		}
	}
}

func (p *parser) parseItem(tok lex.Token) (interface{}, error) {
	switch tok.Val {
	case "tx":
		return p.parseTransaction()
	case "unit":
		return p.parseUnit()
	case "balance":
		return p.parseBalance()
	default:
		return nil, xerrors.Errorf("unknown keyword %v at %v", tok.Val, tok.Pos)
	}
}

func (p *parser) parseTransaction() (interface{}, error) {
	panic(nil)
}

func (p *parser) parseUnit() (interface{}, error) {
	panic(nil)
}

func (p *parser) parseBalance() (balance, error) {
	var b balance
	tok := p.l.NextToken()
	var err error
	b.Date, err = parseDateTok(tok)
	if err != nil {
		return b, xerrors.Errorf("parse balance: %v", err)
	}
	tok = p.l.NextToken()
	if tok.Typ != lex.TokAccount {
		return b, xerrors.Errorf("parse balance: %v", unexpected(tok))
	}
	b.Account = book.Account(tok.Val)
	tok = p.l.NextToken()
	switch tok.Typ {
	case lex.TokDecimal:
		if err := p.parseBalanceSingleAmount(&b, tok); err != nil {
			return b, xerrors.Errorf("parse balance: %v", err)
		}
	case lex.TokNewline:
		if err := p.parseBalanceMultipleAmounts(&b); err != nil {
			return b, xerrors.Errorf("parse balance: %v", err)
		}
	default:
		return b, unexpected(tok)
	}
	return b, nil
}

type balance struct {
	Date    civil.Date
	Account book.Account
	Amounts []book.Amount
}

func (p *parser) parseBalanceSingleAmount(b *balance, tok lex.Token) error {
	d, err := parseDecimalTok(tok)
	if err != nil {
		return err
	}
	unitTok := p.l.NextToken()
	u, err := p.parseUnitTok(unitTok)
	if err != nil {
		return err
	}
	a, err := convertAmount(d, u)
	if err != nil {
		return xerrors.Errorf("at %v, %v", tok.Pos, err)
	}
	b.Amounts = append(b.Amounts, a)
	tok = p.l.NextToken()
	if tok.Typ != lex.TokNewline {
		return unexpected(tok)
	}
	return nil
}

func convertAmount(d decimal, u book.UnitType) (book.Amount, error) {
	if d.scale > u.Scale {
		return book.Amount{}, xerrors.Errorf("amount %v for unit %v divisions too small", d, u)
	}
	return book.Amount{
		Number:   d.number * u.Scale / d.scale,
		UnitType: u,
	}, nil
}

func (p *parser) parseBalanceMultipleAmounts(b *balance) error {
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

func (p *parser) parseUnitTok(tok lex.Token) (book.UnitType, error) {
	if tok.Typ != lex.TokUnit {
		return book.UnitType{}, unexpected(tok)
	}
	u, ok := p.units[tok.Val]
	if !ok {
		return book.UnitType{}, xerrors.Errorf("parse unit %v at %v: unit not declared yet", tok.Val, tok.Pos)
	}
	return u, nil
}

func unexpected(tok lex.Token) error {
	return xerrors.Errorf("unexpected %v token %v at %v", tok.Typ, tok.Val, tok.Pos)
}

func parseDecimalTok(tok lex.Token) (decimal, error) {
	if tok.Typ != lex.TokDecimal {
		return decimal{}, unexpected(tok)
	}
	d, err := parseDecimal(tok.Val)
	if err != nil {
		return d, xerrors.Errorf("parse decimal at %v: %v", tok.Pos, err)
	}
	return d, nil
}

func parseDateTok(tok lex.Token) (civil.Date, error) {
	if tok.Typ != lex.TokDate {
		return civil.Date{}, unexpected(tok)
	}
	d, err := civil.ParseDate(tok.Val)
	if err != nil {
		return d, xerrors.Errorf("parse date at %v: %v", tok.Pos, err)
	}
	return d, nil
}
