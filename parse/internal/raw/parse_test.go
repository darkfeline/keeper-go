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

package raw

import (
	"strings"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
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
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	want := []EntryCommon{
		BalanceEntry{
			Common:  Common{Line: 1},
			Date:    civil.Date{2001, 2, 3},
			Account: "Some:account",
			Amounts: []Amount{
				{Number: Decimal{12345, 100}, Unit: "USD"},
			},
		},
		BalanceEntry{
			Common:  Common{Line: 2},
			Date:    civil.Date{2001, 2, 5},
			Account: "Some:account",
			Amounts: []Amount{
				{Number: Decimal{12345, 100}, Unit: "USD"},
				{Number: Decimal{56700, 1}, Unit: "JPY"},
			},
		},
		UnitEntry{
			Common: Common{Line: 6},
			Symbol: "USD",
			Scale:  Decimal{100, 1},
		},
		TransactionEntry{
			Common:      Common{Line: 7},
			Date:        civil.Date{2001, 2, 3},
			Description: "Buy stuff",
			Splits: []Split{
				{
					Account: "Some:account",
					Amount:  Amount{Number: Decimal{12, 10}, Unit: "USD"},
				},
				{
					Account: "Expenses:Stuff",
					Amount:  Amount{Number: Decimal{-12, 10}, Unit: "USD"},
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestParse_split_without_amount(t *testing.T) {
	t.Parallel()
	const input = `tx 2001-02-03 "Buy stuff"
Some:account 1.2 USD
Expenses:Stuff
.
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	want := []EntryCommon{
		TransactionEntry{
			Common:      Common{Line: 1},
			Date:        civil.Date{2001, 2, 3},
			Description: "Buy stuff",
			Splits: []Split{
				{
					Account: "Some:account",
					Amount:  Amount{Number: Decimal{12, 10}, Unit: "USD"},
				},
				{
					Account: "Expenses:Stuff",
					Amount:  Amount{},
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}
