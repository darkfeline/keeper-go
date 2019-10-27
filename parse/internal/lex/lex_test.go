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

package lex

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLexer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		text string
		want []Token
	}{
		{
			desc: "simple",
			text: `unit USD 100
tx 2001-02-03 "Some description"
Some:account 123.45 USD
Some:account -123.45 USD
.
bal 2001-02-03 Some:account 123.45 USD
`,
			want: []Token{
				{TokKeyword, `unit`, Pos{1, 0}},
				{TokUnit, `USD`, Pos{1, 5}},
				{TokDecimal, `100`, Pos{1, 9}},
				{TokNewline, "\n", Pos{1, 12}},
				{TokKeyword, `tx`, Pos{2, 0}},
				{TokDate, `2001-02-03`, Pos{2, 3}},
				{TokString, `"Some description"`, Pos{2, 14}},
				{TokNewline, "\n", Pos{2, 32}},
				{TokAccount, `Some:account`, Pos{3, 0}},
				{TokDecimal, `123.45`, Pos{3, 13}},
				{TokUnit, `USD`, Pos{3, 20}},
				{TokNewline, "\n", Pos{3, 23}},
				{TokAccount, `Some:account`, Pos{4, 0}},
				{TokDecimal, `-123.45`, Pos{4, 13}},
				{TokUnit, `USD`, Pos{4, 21}},
				{TokNewline, "\n", Pos{4, 24}},
				{TokDot, `.`, Pos{5, 0}},
				{TokNewline, "\n", Pos{5, 1}},
				{TokKeyword, `bal`, Pos{6, 0}},
				{TokDate, `2001-02-03`, Pos{6, 4}},
				{TokAccount, `Some:account`, Pos{6, 15}},
				{TokDecimal, `123.45`, Pos{6, 28}},
				{TokUnit, `USD`, Pos{6, 35}},
				{TokNewline, "\n", Pos{6, 38}},
			},
		},
		{
			desc: "comment",
			text: `tx 2001-02-03 "Some description"  # blah
Some:account 123.45 USD #gascogne is cute
Some:account -123.45 USD
.
`,
			want: []Token{
				{TokKeyword, `tx`, Pos{1, 0}},
				{TokDate, `2001-02-03`, Pos{1, 3}},
				{TokString, `"Some description"`, Pos{1, 14}},
				{TokNewline, "\n", Pos{1, 40}},
				{TokAccount, `Some:account`, Pos{2, 0}},
				{TokDecimal, `123.45`, Pos{2, 13}},
				{TokUnit, `USD`, Pos{2, 20}},
				{TokNewline, "\n", Pos{2, 41}},
				{TokAccount, `Some:account`, Pos{3, 0}},
				{TokDecimal, `-123.45`, Pos{3, 13}},
				{TokUnit, `USD`, Pos{3, 21}},
				{TokNewline, "\n", Pos{3, 24}},
				{TokDot, `.`, Pos{4, 0}},
				{TokNewline, "\n", Pos{4, 1}},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := lexTestString(t, c.text)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("token mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func lexTestString(t *testing.T, s string) []Token {
	t.Helper()
	l := Lex(strings.NewReader(s))
	var got []Token
pump:
	for {
		switch tok := l.NextToken(); tok.Typ {
		case TokEOF:
			break pump
		case TokError:
			t.Fatalf("Lexer returned error: %+v", tok)
		default:
			got = append(got, tok)
		}
	}
	return got
}
