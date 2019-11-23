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
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/raw"
)

func TestParse(t *testing.T) {
	t.Parallel()
	const input = `bal 2001-02-03 Some:account -1.20 USD
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff
.
unit USD 100
`
	got := parseTestInput(t, input)
	u := book.UnitType{Symbol: "USD", Scale: 100}
	want := []interface{}{
		TransactionLine{
			Common: Common{
				Date: civil.Date{2001, 2, 3},
				Line: 2,
			},
			Description: "Buy stuff",
			Splits: []book.Split{
				{
					Account: "Some:account",
					Amount:  book.Amount{Number: -120, UnitType: u},
				},
				{
					Account: "Expenses:Stuff",
					Amount:  book.Amount{Number: 120, UnitType: u},
				},
			},
		},
		BalanceLine{
			Common: Common{
				Date: civil.Date{2001, 2, 3},
				Line: 1,
			},
			Account:  "Some:account",
			Balance:  book.Balance{{Number: -120, UnitType: u}},
			Declared: book.Balance{{Number: -120, UnitType: u}},
		},
	}
	if diff := cmp.Diff(want, got.Lines); diff != "" {
		t.Errorf("transactions mismatch (-want +got):\n%s", diff)
	}
}

func parseTestInput(t *testing.T, input string) Result {
	t.Helper()
	r, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestParse_unbalanced_transaction(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff 1.3 USD
.
`
	got := parseTestInput(t, input)
	if len(got.Errors) == 0 {
		t.Errorf("Expected errors")
	}
}

func TestParse_failed_balance_check(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff
.
bal 2001-02-03 Some:account -1 USD
`
	got := parseTestInput(t, input)
	if len(got.Errors) == 0 {
		t.Errorf("Expected errors")
	}
}

func TestSortEntries(t *testing.T) {
	t.Parallel()
	d1 := civil.Date{2001, 2, 3}
	d2 := civil.Date{2001, 2, 4}
	et1 := raw.TransactionEntry{
		Date:        d1,
		Description: "Buy stuff",
		Splits:      []raw.Split{},
	}
	eb1 := raw.BalanceEntry{
		Date:    d1,
		Account: "Some:account",
		Amounts: []raw.Amount{
			{Number: raw.Decimal{12345, 100}, Unit: "USD"},
		},
	}
	eb2 := raw.BalanceEntry{
		Date:    d2,
		Account: "Some:account",
		Amounts: []raw.Amount{
			{Number: raw.Decimal{22345, 100}, Unit: "USD"},
		},
	}
	eu := raw.UnitEntry{
		Symbol: "USD",
		Scale:  raw.Decimal{100, 1},
	}

	got := []interface{}{et1, eb2, eb1, eu}
	sortEntries(got)
	want := []interface{}{eu, et1, eb1, eb2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestCombineDecimalUnit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    raw.Decimal
		u    book.UnitType
		want int64
	}{
		{raw.Decimal{5, 1000}, book.UnitType{Symbol: "Foo", Scale: 1000}, 5},
		{raw.Decimal{5, 10}, book.UnitType{Symbol: "Foo", Scale: 1000}, 500},
		{raw.Decimal{50, 10}, book.UnitType{Symbol: "Foo", Scale: 1}, 5},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%v %v", c.d, c.u), func(t *testing.T) {
			t.Parallel()
			got, err := combineDecimalUnit(c.d, c.u)
			if err != nil {
				t.Error(err)
			}
			want := book.Amount{
				Number:   c.want,
				UnitType: c.u,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("amount mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsPower10(t *testing.T) {
	t.Parallel()
	cases := []struct {
		n    int64
		want bool
	}{
		{0, false},
		{11, false},
		{-11, false},
		{101, false},
		{-101, false},
		{1, true},
		{-1, true},
		{10, true},
		{100, true},
		{-10, true},
		{-100, true},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d", c.n), func(t *testing.T) {
			t.Parallel()
			got := isPower10(c.n)
			if got != c.want {
				t.Errorf("isPower10(%d) = %v; want %v", c.n, got, c.want)
			}
		})
	}
}
