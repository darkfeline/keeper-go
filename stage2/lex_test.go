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

package stage2

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
		want []token
	}{
		{
			desc: "simple",
			text: `tx 2001-02-03 "Some description"
some:account 123.45 USD
some:account -123.45 USD
.
bal 2001-02-03E4 some:account 123.45 USD
`,
			want: []token{
				{tokKeyword, `tx`},
				{tokOrdering, `2001-02-03`},
				{tokString, `"Some description"`},
				{tokNewline, "\n"},
				{tokAccount, `some:account`},
				{tokDecimal, `123.45`},
				{tokUnit, `USD`},
				{tokNewline, "\n"},
				{tokAccount, `some:account`},
				{tokDecimal, `-123.45`},
				{tokUnit, `USD`},
				{tokNewline, "\n"},
				{tokDot, `.`},
				{tokNewline, "\n"},
				{tokKeyword, `bal`},
				{tokOrdering, `2001-02-03E4`},
				{tokAccount, `some:account`},
				{tokDecimal, `123.45`},
				{tokUnit, `USD`},
				{tokNewline, "\n"},
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

func lexTestString(t *testing.T, s string) []token {
	t.Helper()
	l := lex(strings.NewReader(s))
	var got []token
pump:
	for {
		switch tok := l.nextToken(); tok.Typ {
		case tokEOF:
			break pump
		case tokError:
			t.Fatalf("Lexer returned error: %+v", tok)
		default:
			got = append(got, tok)
		}
	}
	return got
}
