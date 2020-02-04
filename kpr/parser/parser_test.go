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
	const input = `bal 2001-02-03 Some:account 123.45 USD
bal 2001-02-05 Some:account
123.45 USD
56700 JPY
.
unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account 1.2 USD
Expenses:Stuff -1.2 USD
.
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		ast.SingleBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  1,
				Date:    val(5, token.DATE, "2001-02-03"),
				Account: val(16, token.ACCOUNT, "Some:account"),
			},
			Amount: ast.Amount{
				Decimal: val(29, token.DECIMAL, "123.45"),
				Unit:    val(36, token.IDENT, "USD"),
			},
		},
		ast.MultiBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  40,
				Date:    val(44, token.DATE, "2001-02-05"),
				Account: val(55, token.ACCOUNT, "Some:account"),
			},
			Amounts: []ast.LineNode{
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(68, token.DECIMAL, "123.45"),
						Unit:    val(75, token.IDENT, "USD"),
					},
				},
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(79, token.DECIMAL, "56700"),
						Unit:    val(85, token.IDENT, "JPY"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 89},
		},
		ast.UnitDecl{
			TokPos: 91,
			Unit:   val(96, token.IDENT, "USD"),
			Scale:  val(100, token.DECIMAL, "100"),
		},
		ast.Transaction{
			TokPos:      104,
			Date:        val(107, token.DATE, "2001-02-03"),
			Description: val(118, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				ast.Split{
					Account: val(130, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(143, token.DECIMAL, "1.2"),
						Unit:    val(147, token.IDENT, "USD"),
					},
				},
				ast.Split{
					Account: val(151, token.ACCOUNT, "Expenses:Stuff"),
					Amount: &ast.Amount{
						Decimal: val(166, token.DECIMAL, "-1.2"),
						Unit:    val(171, token.IDENT, "USD"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 175},
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
.
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		ast.Transaction{
			TokPos:      1,
			Date:        val(4, token.DATE, "2001-02-03"),
			Description: val(15, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				ast.Split{
					Account: val(27, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(40, token.DECIMAL, "1.2"),
						Unit:    val(44, token.IDENT, "USD"),
					},
				},
				ast.Split{
					Account: val(48, token.ACCOUNT, "Expenses:Stuff"),
				},
			},
			Dot: ast.Dot{TokPos: 63},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParseBytes_empty_lines_ignored(t *testing.T) {
	t.Parallel()
	const input = `
bal 2001-02-05 Some:account
123.45 USD
# some comment
56700 JPY
.
# some comment
unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account 1.2 USD
# some comment
Expenses:Stuff -1.2 USD
.
`
	got, err := ParseBytes(token.NewFileSet(), "", []byte(input), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []ast.Entry{
		ast.MultiBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  2,
				Date:    val(6, token.DATE, "2001-02-05"),
				Account: val(17, token.ACCOUNT, "Some:account"),
			},
			Amounts: []ast.LineNode{
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(30, token.DECIMAL, "123.45"),
						Unit:    val(37, token.IDENT, "USD"),
					},
				},
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(56, token.DECIMAL, "56700"),
						Unit:    val(62, token.IDENT, "JPY"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 66},
		},
		ast.UnitDecl{
			TokPos: 83,
			Unit:   val(88, token.IDENT, "USD"),
			Scale:  val(92, token.DECIMAL, "100"),
		},
		ast.Transaction{
			TokPos:      96,
			Date:        val(99, token.DATE, "2001-02-03"),
			Description: val(110, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				ast.Split{
					Account: val(122, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(135, token.DECIMAL, "1.2"),
						Unit:    val(139, token.IDENT, "USD"),
					},
				},
				ast.Split{
					Account: val(158, token.ACCOUNT, "Expenses:Stuff"),
					Amount: &ast.Amount{
						Decimal: val(173, token.DECIMAL, "-1.2"),
						Unit:    val(178, token.IDENT, "USD"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 182},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func val(pos token.Pos, tok token.Token, lit string) ast.BasicValue {
	return ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}
}
