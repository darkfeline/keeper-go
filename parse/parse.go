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
			return nil, xerrors.Errorf("keeper/parse: lex error: %v", tok.Val)
		default:
			return nil, xerrors.Errorf("keeper/parse: unexpected token: %v", tok.Val)
		}
	}
}

type balance struct {
	Date    civil.Date
	Account book.Account
	Amounts []book.Amount
}
