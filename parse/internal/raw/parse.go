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

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/internal/lex"
)

func Parse(r io.Reader) ([]interface{}, error) {
	panic("Not implemented")
}

type parser struct {
	l *lex.Lexer
}

func newParser(r io.Reader) *parser {
	return &parser{
		l: lex.Lex(r),
	}
}

func (p *parser) parse(r io.Reader) ([]interface{}, error) {
	var items []interface{}
	for {
		switch tok := p.l.NextToken(); tok.Typ {
		case lex.TokEOF:
			return items, nil
		case lex.TokError:
			return items, fmt.Errorf("raw: lex error: %v at %v", tok.Val, tok.Pos)
		case lex.TokKeyword:
			v, err := p.parseItem(tok)
			if err != nil {
				return nil, fmt.Errorf("raw: %v", err)
			}
			items = append(items, v)
		default:
			return nil, fmt.Errorf("raw: %v", unexpected(tok))
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
		return nil, fmt.Errorf("unknown keyword %v at %v", tok.Val, tok.Pos)
	}
}

func (p *parser) parseTransaction() (interface{}, error) {
	panic(nil)
}

func (p *parser) parseUnit() (interface{}, error) {
	panic(nil)
}

func (p *parser) parseBalance() (Balance, error) {
	var b Balance
	var err error

	tok := p.l.NextToken()
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

func (p *parser) parseBalanceSingleAmount(b *Balance, tok lex.Token) error {
	d, err := parseDecimalTok(tok)
	if err != nil {
		return err
	}

	unitTok := p.l.NextToken()
	b.Amounts = append(b.Amounts, Amount{
		Number: d,
		Unit:   unitTok.Val,
	})

	tok = p.l.NextToken()
	if tok.Typ != lex.TokNewline {
		return unexpected(tok)
	}
	return nil
}

func (p *parser) parseBalanceMultipleAmounts(b *Balance) error {
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

func unexpected(tok lex.Token) error {
	return fmt.Errorf("unexpected %v token %v at %v", tok.Typ, tok.Val, tok.Pos)
}

func parseDecimalTok(tok lex.Token) (Decimal, error) {
	if tok.Typ != lex.TokDecimal {
		return Decimal{}, unexpected(tok)
	}
	d, err := parseDecimal(tok.Val)
	if err != nil {
		return d, fmt.Errorf("parse decimal at %v: %v", tok.Pos, err)
	}
	return d, nil
}

func parseDateTok(tok lex.Token) (civil.Date, error) {
	if tok.Typ != lex.TokDate {
		return civil.Date{}, unexpected(tok)
	}
	d, err := civil.ParseDate(tok.Val)
	if err != nil {
		return d, fmt.Errorf("parse date at %v: %v", tok.Pos, err)
	}
	return d, nil
}
