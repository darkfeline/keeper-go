// Copyright (C) 2020  Allen Li
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

package journal

import (
	"fmt"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/kpr/token"
)

func TestBuildEntries_simple(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
balance 2001-02-03 Some:account -1.20 USD
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff
end
`
	got, err := buildEntries([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	u := Unit{Symbol: "USD", Scale: 100}
	want := []Entry{
		BalanceAssert{
			EntryDate: civil.Date{2001, 2, 3},
			EntryPos:  token.Position{Offset: 13, Line: 2, Column: 1},
			Account:   "Some:account",
			Declared:  Balance{{Number: -120, Unit: u}},
		},
		Transaction{
			EntryDate:   civil.Date{2001, 2, 3},
			EntryPos:    token.Position{Offset: 55, Line: 3, Column: 1},
			Description: "Buy stuff",
			Splits: []Split{
				{
					Account: "Some:account",
					Amount:  Amount{Number: -120, Unit: u},
				},
				{
					Account: "Expenses:Stuff",
					Amount:  Amount{Number: 120, Unit: u},
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("transactions mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildEntries_unbalanced(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff 1.3 USD
end
`
	_, err := buildEntries([]byte(input))
	if err == nil {
		t.Errorf("Expected errors")
	}
}

func TestCombineDecimalUnit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    decimal
		u    Unit
		want int64
	}{
		{decimal{5, 1000}, Unit{Symbol: "Foo", Scale: 1000}, 5},
		{decimal{5, 10}, Unit{Symbol: "Foo", Scale: 1000}, 500},
		{decimal{50, 10}, Unit{Symbol: "Foo", Scale: 1}, 5},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%v %v", c.d, c.u), func(t *testing.T) {
			t.Parallel()
			got, err := combineDecimalUnit(c.d, c.u)
			if err != nil {
				t.Error(err)
			}
			want := Amount{
				Number: c.want,
				Unit:   c.u,
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
