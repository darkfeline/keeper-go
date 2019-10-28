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
	"go.felesatra.moe/keeper/parse/internal/raw"
)

func TestParse(t *testing.T) {
	t.Parallel()
	const input = `bal 2001-02-03 Some:account -1.20 USD
unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff
.
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	want := []book.Transaction{
		{
			Date:        civil.Date{2001, 2, 3},
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
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("transactions mismatch (-want +got):\n%s", diff)
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

	got := []raw.EntryCommon{et1, eb2, eb1, eu}
	sortEntries(got)
	want := []raw.EntryCommon{eu, et1, eb1, eb2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestConvertAmount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    raw.Decimal
		u    book.UnitType
		want int64
	}{
		{raw.Decimal{5, 1000}, book.UnitType{Symbol: "Foo", Scale: 1000}, 5},
		{raw.Decimal{5, 10}, book.UnitType{Symbol: "Foo", Scale: 1000}, 500},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%v %v", c.d, c.u), func(t *testing.T) {
			t.Parallel()
			got, err := convertAmount(c.d, &c.u)
			if err != nil {
				t.Error(err)
			}
			want := book.Amount{
				Number:   c.want,
				UnitType: &c.u,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("amount mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
