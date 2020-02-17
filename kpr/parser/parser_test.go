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
				Date:    val(9, token.DATE, "2001-02-03"),
				Account: val(20, token.ACCOUNT, "Some:account"),
			},
			Amount: ast.Amount{
				Decimal: val(33, token.DECIMAL, "123.45"),
				Unit:    val(40, token.UNIT_SYM, "USD"),
			},
		},
		ast.MultiBalance{
			BalanceHeader: ast.BalanceHeader{
				TokPos:  44,
				Date:    val(52, token.DATE, "2001-02-05"),
				Account: val(63, token.ACCOUNT, "Some:account"),
			},
			Amounts: []ast.LineNode{
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(76, token.DECIMAL, "123.45"),
						Unit:    val(83, token.UNIT_SYM, "USD"),
					},
				},
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(87, token.DECIMAL, "56700"),
						Unit:    val(93, token.UNIT_SYM, "JPY"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 97},
		},
		ast.UnitDecl{
			TokPos: 99,
			Unit:   val(104, token.UNIT_SYM, "USD"),
			Scale:  val(108, token.DECIMAL, "100"),
		},
		ast.Transaction{
			TokPos:      112,
			Date:        val(115, token.DATE, "2001-02-03"),
			Description: val(126, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				ast.Split{
					Account: val(138, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(151, token.DECIMAL, "1.2"),
						Unit:    val(155, token.UNIT_SYM, "USD"),
					},
				},
				ast.Split{
					Account: val(159, token.ACCOUNT, "Expenses:Stuff"),
					Amount: &ast.Amount{
						Decimal: val(174, token.DECIMAL, "-1.2"),
						Unit:    val(179, token.UNIT_SYM, "USD"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 183},
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
						Unit:    val(44, token.UNIT_SYM, "USD"),
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
balance 2001-02-05 Some:account
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
				Date:    val(10, token.DATE, "2001-02-05"),
				Account: val(21, token.ACCOUNT, "Some:account"),
			},
			Amounts: []ast.LineNode{
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(34, token.DECIMAL, "123.45"),
						Unit:    val(41, token.UNIT_SYM, "USD"),
					},
				},
				ast.AmountLine{
					Amount: ast.Amount{
						Decimal: val(60, token.DECIMAL, "56700"),
						Unit:    val(66, token.UNIT_SYM, "JPY"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 70},
		},
		ast.UnitDecl{
			TokPos: 87,
			Unit:   val(92, token.UNIT_SYM, "USD"),
			Scale:  val(96, token.DECIMAL, "100"),
		},
		ast.Transaction{
			TokPos:      100,
			Date:        val(103, token.DATE, "2001-02-03"),
			Description: val(114, token.STRING, `"Buy stuff"`),
			Splits: []ast.LineNode{
				ast.Split{
					Account: val(126, token.ACCOUNT, "Some:account"),
					Amount: &ast.Amount{
						Decimal: val(139, token.DECIMAL, "1.2"),
						Unit:    val(143, token.UNIT_SYM, "USD"),
					},
				},
				ast.Split{
					Account: val(162, token.ACCOUNT, "Expenses:Stuff"),
					Amount: &ast.Amount{
						Decimal: val(177, token.DECIMAL, "-1.2"),
						Unit:    val(182, token.UNIT_SYM, "USD"),
					},
				},
			},
			Dot: ast.Dot{TokPos: 186},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func val(pos token.Pos, tok token.Token, lit string) ast.BasicValue {
	return ast.BasicValue{ValuePos: pos, Kind: tok, Value: lit}
}
