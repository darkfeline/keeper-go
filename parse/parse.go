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

func parse(r io.Reader) ([]interface{}, error) {
	var items []interface{}
	l := lex.Lex(r)
	for {
		switch tok := l.NextToken(); tok.Typ {
		case lex.TokEOF:
			return items, nil
		case lex.TokError:
			return nil, xerrors.Errorf("keeper/parse: lex error: %v at %v", tok.Val, tok.Pos)
		case lex.TokKeyword:
			v, err := parseItem(l, tok)
			if err != nil {
				return nil, xerrors.Errorf("keeper/parse: %v", err)
			}
			items = append(items, v)
		default:
			return nil, xerrors.Errorf("keeper/parse: unexpected token: %v at %v", tok.Val, tok.Pos)
		}
	}
}

func parseItem(l *lex.Lexer, tok lex.Token) (interface{}, error) {
	switch tok.Val {
	case "tx":
		return parseTransaction(l)
	case "unit":
		return parseUnit(l)
	case "balance":
		return parseBalance(l)
	default:
		return nil, xerrors.Errorf("unknown keyword %v at %v", tok.Val, tok.Pos)
	}
}

func parseTransaction(l *lex.Lexer) (interface{}, error) {
	panic(nil)
}

func parseUnit(l *lex.Lexer) (interface{}, error) {
	panic(nil)
}

func parseBalance(l *lex.Lexer) (balance, error) {
	var b balance
	tok := l.NextToken()
	if err := expect(tok, lex.TokDate); err != nil {
		return b, xerrors.Errorf("parse balance: %v", err)
	}
	var err error
	b.Date, err = parseDate(tok)
	if err != nil {
		return b, xerrors.Errorf("parse balance: %v", err)
	}
	tok = l.NextToken()
	if err := expect(tok, lex.TokAccount); err != nil {
		return b, xerrors.Errorf("parse balance: %v", err)
	}
	b.Account = book.Account(tok.Val)
	// XXXXXXXXXXX single and multi
	return b, nil
}

func expect(tok lex.Token, t lex.TokenType) error {
	if tok.Typ != t {
		return xerrors.Errorf("unexpected token %v at %v", tok.Val, tok.Pos)
	}
	return nil
}

func parseDate(tok lex.Token) (civil.Date, error) {
	if tok.Typ != lex.TokDate {
		panic(tok)
	}
	d, err := civil.ParseDate(tok.Val)
	if err != nil {
		return d, xerrors.Errorf("parse date at %v: %v", tok.Pos, err)
	}
	return d, nil
}

type balance struct {
	Date    civil.Date
	Account book.Account
	Amounts []book.Amount
}
