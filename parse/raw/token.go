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

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/parse/internal/lex"
)

func unexpected(tok lex.Token) error {
	return fmt.Errorf("unexpected %v", tok)
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

func parseUnitTok(tok lex.Token) (string, error) {
	if tok.Typ != lex.TokUnit {
		return "", unexpected(tok)
	}
	return tok.Val, nil
}

func parseStringTok(tok lex.Token) (string, error) {
	if tok.Typ != lex.TokString {
		return "", unexpected(tok)
	}
	n := len(tok.Val)
	if tok.Val[0] != '"' || tok.Val[n-1] != '"' {
		panic("string token starts without double quotes")
	}
	var s []rune
	var escaped bool
	for _, r := range tok.Val[1 : n-1] {
		if r == '\\' && !escaped {
			escaped = true
			continue
		}
		s = append(s, r)
		escaped = false
	}
	return string(s), nil
}