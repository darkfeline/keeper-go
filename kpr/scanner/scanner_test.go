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

package scanner

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/kpr/token"
)

func TestLexer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		text string
		mode Mode
		want []result
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
			want: []result{
				{1, token.IDENT, `unit`},
				{6, token.IDENT, `USD`},
				{10, token.DECIMAL, `100`},
				{13, token.NEWLINE, "\n"},
				{14, token.IDENT, `tx`},
				{17, token.DATE, `2001-02-03`},
				{28, token.STRING, `"Some description"`},
				{46, token.NEWLINE, "\n"},
				{47, token.ACCOUNT, `Some:account`},
				{60, token.DECIMAL, `123.45`},
				{67, token.IDENT, `USD`},
				{70, token.NEWLINE, "\n"},
				{71, token.ACCOUNT, `Some:account`},
				{84, token.DECIMAL, `-123.45`},
				{92, token.IDENT, `USD`},
				{95, token.NEWLINE, "\n"},
				{96, token.DOT, `.`},
				{97, token.NEWLINE, "\n"},
				{98, token.IDENT, `bal`},
				{102, token.DATE, `2001-02-03`},
				{113, token.ACCOUNT, `Some:account`},
				{126, token.DECIMAL, `123.45`},
				{133, token.IDENT, `USD`},
				{136, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "comment ignored",
			text: `tx 2001-02-03 "Some description"  # blah
Some:account 123.45 USD #gascogne is cute
Some:account -123.45 USD
.
`,
			want: []result{
				{1, token.IDENT, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{41, token.NEWLINE, "\n"},
				{42, token.ACCOUNT, `Some:account`},
				{55, token.DECIMAL, `123.45`},
				{62, token.IDENT, `USD`},
				{83, token.NEWLINE, "\n"},
				{84, token.ACCOUNT, `Some:account`},
				{97, token.DECIMAL, `-123.45`},
				{105, token.IDENT, `USD`},
				{108, token.NEWLINE, "\n"},
				{109, token.DOT, `.`},
				{110, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "comment tokenized",
			text: `tx 2001-02-03 "Some description"  # blah
Some:account 123.45 USD #gascogne is cute
Some:account -123.45 USD
.
`,
			mode: ScanComments,
			want: []result{
				{1, token.IDENT, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{35, token.COMMENT, `# blah`},
				{41, token.NEWLINE, "\n"},
				{42, token.ACCOUNT, `Some:account`},
				{55, token.DECIMAL, `123.45`},
				{62, token.IDENT, `USD`},
				{66, token.COMMENT, `#gascogne is cute`},
				{83, token.NEWLINE, "\n"},
				{84, token.ACCOUNT, `Some:account`},
				{97, token.DECIMAL, `-123.45`},
				{105, token.IDENT, `USD`},
				{108, token.NEWLINE, "\n"},
				{109, token.DOT, `.`},
				{110, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "account with number",
			text: `tx 2001-02-03 "Some description"
Some:account4 123.45 USD
.
`,
			want: []result{
				{1, token.IDENT, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{33, token.NEWLINE, "\n"},
				{34, token.ACCOUNT, `Some:account4`},
				{48, token.DECIMAL, `123.45`},
				{55, token.IDENT, `USD`},
				{58, token.NEWLINE, "\n"},
				{59, token.DOT, `.`},
				{60, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "decimal with comma",
			text: `tx 2001-02-03 "Some description"
Some:account 2,123.45 USD
.
`,
			want: []result{
				{1, token.IDENT, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{33, token.NEWLINE, "\n"},
				{34, token.ACCOUNT, `Some:account`},
				{47, token.DECIMAL, `2,123.45`},
				{56, token.IDENT, `USD`},
				{59, token.NEWLINE, "\n"},
				{60, token.DOT, `.`},
				{61, token.NEWLINE, "\n"},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := lexTestString(t, c.text, c.mode)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("token mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func lexTestString(t *testing.T, src string, mode Mode) []result {
	t.Helper()
	fs := token.NewFileSet()
	f := fs.AddFile("", -1, len(src))
	var s Scanner
	var errs ErrorList
	s.Init(f, []byte(src), errs.Add, mode)
	var got []result
pump:
	for {
		switch pos, tok, lit := s.Scan(); tok {
		case token.EOF:
			break pump
		default:
			got = append(got, result{
				Pos: pos,
				Tok: tok,
				Lit: lit,
			})
		}
	}
	if s.ErrorCount != 0 {
		t.Errorf("scanner has non-zero ErrorCount %d: %s",
			s.ErrorCount, errs)
	}
	return got
}
