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

package parser

import (
	"testing"

	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/token"

	"github.com/google/go-cmp/cmp"
)

func TestParseBytes(t *testing.T) {
	t.Parallel()
	const input = `balance 2001-02-03 Some:account 123.45 USD
balance 2001-02-05 Some:account
123.45 USD
56700 JPY
end
unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account 1.2 USD
Expenses:Stuff -1.2 USD
end
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		&ast.SingleBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  1,
				Date:    val(9, token.DATE, "2001-02-03"),
				Account: val(20, token.ACCOUNT, "Some:account"),
			},
			Amount: &ast.Amount{
				Decimal: val(33, token.DECIMAL, "123.45"),
				Unit:    val(40, token.USYMBOL, "USD"),
			},
		},
		&ast.MultiBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  44,
				Date:    val(52, token.DATE, "2001-02-05"),
				Account: val(63, token.ACCOUNT, "Some:account"),
			},
			Amounts: []ast.LineNode{
				&ast.AmountLine{Amount: &ast.Amount{
					Decimal: val(76, token.DECIMAL, "123.45"),
					Unit:    val(83, token.USYMBOL, "USD"),
				}},
				&ast.AmountLine{Amount: &ast.Amount{
					Decimal: val(87, token.DECIMAL, "56700"),
					Unit:    val(93, token.USYMBOL, "JPY"),
				}},
			},
			EndTok: &ast.End{TokPos: 97},
		},
		&ast.UnitDecl{
			TokPos: 101,
			Unit:   val(106, token.USYMBOL, "USD"),
			Scale:  val(110, token.DECIMAL, "100"),
		},
		&ast.Transaction{
			TokPos:      114,
			Date:        val(117, token.DATE, "2001-02-03"),
			Description: val(128, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				&ast.SplitLine{
					Account: val(140, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(153, token.DECIMAL, "1.2"),
						Unit:    val(157, token.USYMBOL, "USD"),
					},
				},
				&ast.SplitLine{
					Account: val(161, token.ACCOUNT, "Expenses:Stuff"),
					Amount: &ast.Amount{
						Decimal: val(176, token.DECIMAL, "-1.2"),
						Unit:    val(181, token.USYMBOL, "USD"),
					},
				},
			},
			EndTok: &ast.End{TokPos: 185},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_invalid_token(t *testing.T) {
	t.Parallel()
	const input = `.`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err == nil {
		t.Errorf("Expected error")
	}
	want := []ast.Entry{
		&ast.BadEntry{
			From: 1,
			To:   2,
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_empty_lines_ignored(t *testing.T) {
	t.Parallel()
	const input = `
balance 2001-02-05 Some:account
123.45 USD
# some comment
56700 JPY
end
# some comment
unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account 1.2 USD
# some comment
Expenses:Stuff -1.2 USD
end
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		&ast.MultiBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  2,
				Date:    val(10, token.DATE, "2001-02-05"),
				Account: val(21, token.ACCOUNT, "Some:account"),
			},
			Amounts: []ast.LineNode{
				&ast.AmountLine{Amount: &ast.Amount{
					Decimal: val(34, token.DECIMAL, "123.45"),
					Unit:    val(41, token.USYMBOL, "USD"),
				}},
				&ast.AmountLine{Amount: &ast.Amount{
					Decimal: val(60, token.DECIMAL, "56700"),
					Unit:    val(66, token.USYMBOL, "JPY"),
				}},
			},
			EndTok: &ast.End{TokPos: 70},
		},
		&ast.UnitDecl{
			TokPos: 89,
			Unit:   val(94, token.USYMBOL, "USD"),
			Scale:  val(98, token.DECIMAL, "100"),
		},
		&ast.Transaction{
			TokPos:      102,
			Date:        val(105, token.DATE, "2001-02-03"),
			Description: val(116, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				&ast.SplitLine{
					Account: val(128, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(141, token.DECIMAL, "1.2"),
						Unit:    val(145, token.USYMBOL, "USD"),
					},
				},
				&ast.SplitLine{
					Account: val(164, token.ACCOUNT, "Expenses:Stuff"),
					Amount: &ast.Amount{
						Decimal: val(179, token.DECIMAL, "-1.2"),
						Unit:    val(184, token.USYMBOL, "USD"),
					},
				},
			},
			EndTok: &ast.End{TokPos: 188},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_split_without_amount(t *testing.T) {
	t.Parallel()
	const input = `tx 2001-02-03 "Buy stuff"
Some:account 1.2 USD
Expenses:Stuff
end
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		&ast.Transaction{
			TokPos:      1,
			Date:        val(4, token.DATE, "2001-02-03"),
			Description: val(15, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				&ast.SplitLine{
					Account: val(27, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(40, token.DECIMAL, "1.2"),
						Unit:    val(44, token.USYMBOL, "USD"),
					},
				},
				&ast.SplitLine{
					Account: val(48, token.ACCOUNT, "Expenses:Stuff"),
				},
			},
			EndTok: &ast.End{TokPos: 63},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_unterminated_tx(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff 1.2 USD
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err == nil {
		t.Errorf("Expected error")
	}
	want := []ast.Entry{
		&ast.UnitDecl{
			TokPos: 1,
			Unit:   val(6, token.USYMBOL, "USD"),
			Scale:  val(10, token.DECIMAL, "100"),
		},
		&ast.BadEntry{From: 14, To: 85},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_truncated_split(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err == nil {
		t.Errorf("Expected error")
	}
	want := []ast.Entry{
		&ast.UnitDecl{
			TokPos: 1,
			Unit:   val(6, token.USYMBOL, "USD"),
			Scale:  val(10, token.DECIMAL, "100"),
		},
		&ast.BadEntry{From: 14, To: 53},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_close(t *testing.T) {
	t.Parallel()
	const input = `close 2001-02-03 Some:account
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		&ast.CloseAccount{
			TokPos:  1,
			Date:    val(7, token.DATE, "2001-02-03"),
			Account: val(18, token.ACCOUNT, "Some:account"),
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func val(pos token.Pos, tok token.Token, lit string) *ast.BasicValue {
	return &ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}
}
