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

func TestScanner(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		text string
		mode Mode
		want []result
	}{
		{
			desc: "full example",
			text: `unit USD 100
tx 2001-02-03 "Some description"
Some:account 123.45 USD
Some:account -123.45 USD
end
balance 2001-02-03 Some:account 123.45 USD
account Some:account
"foo" "bar"
end
`,
			want: []result{
				{1, token.UNIT, `unit`},
				{6, token.USYMBOL, `USD`},
				{10, token.DECIMAL, `100`},
				{13, token.NEWLINE, "\n"},
				{14, token.TX, `tx`},
				{17, token.DATE, `2001-02-03`},
				{28, token.STRING, `"Some description"`},
				{46, token.NEWLINE, "\n"},
				{47, token.ACCTNAME, `Some:account`},
				{60, token.DECIMAL, `123.45`},
				{67, token.USYMBOL, `USD`},
				{70, token.NEWLINE, "\n"},
				{71, token.ACCTNAME, `Some:account`},
				{84, token.DECIMAL, `-123.45`},
				{92, token.USYMBOL, `USD`},
				{95, token.NEWLINE, "\n"},
				{96, token.END, `end`},
				{99, token.NEWLINE, "\n"},
				{100, token.BALANCE, `balance`},
				{108, token.DATE, `2001-02-03`},
				{119, token.ACCTNAME, `Some:account`},
				{132, token.DECIMAL, `123.45`},
				{139, token.USYMBOL, `USD`},
				{142, token.NEWLINE, "\n"},
				{143, token.ACCOUNT, "account"},
				{151, token.ACCTNAME, "Some:account"},
				{163, token.NEWLINE, "\n"},
				{164, token.STRING, `"foo"`},
				{170, token.STRING, `"bar"`},
				{175, token.NEWLINE, "\n"},
				{176, token.END, "end"},
				{179, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "comment ignored",
			text: `tx 2001-02-03 "Some description"  # blah
Some:account 123.45 USD #gascogne is cute
Some:account -123.45 USD
end
`,
			want: []result{
				{1, token.TX, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{41, token.NEWLINE, "\n"},
				{42, token.ACCTNAME, `Some:account`},
				{55, token.DECIMAL, `123.45`},
				{62, token.USYMBOL, `USD`},
				{83, token.NEWLINE, "\n"},
				{84, token.ACCTNAME, `Some:account`},
				{97, token.DECIMAL, `-123.45`},
				{105, token.USYMBOL, `USD`},
				{108, token.NEWLINE, "\n"},
				{109, token.END, `end`},
				{112, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "comment tokenized",
			text: `tx 2001-02-03 "Some description"  # blah
Some:account 123.45 USD #gascogne is cute
Some:account -123.45 USD
end
`,
			mode: ScanComments,
			want: []result{
				{1, token.TX, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{35, token.COMMENT, `# blah`},
				{41, token.NEWLINE, "\n"},
				{42, token.ACCTNAME, `Some:account`},
				{55, token.DECIMAL, `123.45`},
				{62, token.USYMBOL, `USD`},
				{66, token.COMMENT, `#gascogne is cute`},
				{83, token.NEWLINE, "\n"},
				{84, token.ACCTNAME, `Some:account`},
				{97, token.DECIMAL, `-123.45`},
				{105, token.USYMBOL, `USD`},
				{108, token.NEWLINE, "\n"},
				{109, token.END, `end`},
				{112, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "account with number",
			text: `tx 2001-02-03 "Some description"
Some:account4 123.45 USD
end
`,
			want: []result{
				{1, token.TX, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{33, token.NEWLINE, "\n"},
				{34, token.ACCTNAME, `Some:account4`},
				{48, token.DECIMAL, `123.45`},
				{55, token.USYMBOL, `USD`},
				{58, token.NEWLINE, "\n"},
				{59, token.END, `end`},
				{62, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "account with underscore",
			text: `Some:account_4`,
			want: []result{
				{1, token.ACCTNAME, `Some:account_4`},
			},
		},
		{
			desc: "decimal with comma",
			text: `tx 2001-02-03 "Some description"
Some:account 2,123.45 USD
end
`,
			want: []result{
				{1, token.TX, `tx`},
				{4, token.DATE, `2001-02-03`},
				{15, token.STRING, `"Some description"`},
				{33, token.NEWLINE, "\n"},
				{34, token.ACCTNAME, `Some:account`},
				{47, token.DECIMAL, `2,123.45`},
				{56, token.USYMBOL, `USD`},
				{59, token.NEWLINE, "\n"},
				{60, token.END, `end`},
				{63, token.NEWLINE, "\n"},
			},
		},
		{
			desc: "string escape",
			text: `"\"\\" USD`,
			want: []result{
				{1, token.STRING, `"\"\\"`},
				{8, token.USYMBOL, `USD`},
			},
		},
		{
			desc: "disable",
			text: `disable 2001-02-03 Some:account`,
			want: []result{
				{1, token.DISABLE, "disable"},
				{9, token.DATE, "2001-02-03"},
				{20, token.ACCTNAME, "Some:account"},
			},
		},
		{
			desc: "account",
			text: `account Some:account
"foo" "bar"
end
`,
			want: []result{
				{1, token.ACCOUNT, "account"},
				{9, token.ACCTNAME, "Some:account"},
				{21, token.NEWLINE, "\n"},
				{22, token.STRING, `"foo"`},
				{28, token.STRING, `"bar"`},
				{33, token.NEWLINE, "\n"},
				{34, token.END, "end"},
				{37, token.NEWLINE, "\n"},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			s, got, errs := scanString(c.text, c.mode)
			if s.ErrorCount != 0 {
				t.Errorf("scanner has non-zero ErrorCount %d: %s",
					s.ErrorCount, errs)
			}
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("token mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestScanner_errors(t *testing.T) {
	t.Parallel()
	const src = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff 1.2 USD
.
`
	s, got, _ := scanString(src, 0)
	if s.ErrorCount == 0 {
		t.Errorf("Expected errors")
		t.Logf("Got tokens: %+v", got)
	}
}

func TestScanner_bad_token(t *testing.T) {
	t.Parallel()
	const src = `.`
	s, got, _ := scanString(src, 0)
	if s.ErrorCount == 0 {
		t.Errorf("Expected errors")
	}
	want := []result{
		{1, token.ILLEGAL, `.`},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("token mismatch (-want +got):\n%s", diff)
	}
}

func scanString(src string, mode Mode) (Scanner, []result, ErrorList) {
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
	return s, got, errs
}
