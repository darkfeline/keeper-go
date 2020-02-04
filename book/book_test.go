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

package book

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/journal"
)

func TestCompile(t *testing.T) {
	t.Parallel()
	u := journal.Unit{Symbol: "USD", Scale: 100}
	e := []journal.Entry{
		journal.BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Balance:   journal.Balance{{Number: -232, Unit: u}},
		},
		journal.Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "buy stuff",
			Splits: []journal.Split{
				{
					Account: "Assets:Cash",
					Amount:  journal.Amount{Number: -123, Unit: u},
				},
				{
					Account: "Expenses:Food",
					Amount:  journal.Amount{Number: 123, Unit: u},
				},
			},
		},
	}
	got := Compile(e)
	t.Run("entries", func(t *testing.T) {
		want := []Entry{
			Transaction{
				EntryDate:   civil.Date{2000, 1, 2},
				Description: "buy stuff",
				Splits: []journal.Split{
					{
						Account: "Assets:Cash",
						Amount:  journal.Amount{Number: -123, Unit: u},
					},
					{
						Account: "Expenses:Food",
						Amount:  journal.Amount{Number: 123, Unit: u},
					},
				},
				Balances: TBalance{
					"Assets:Cash":   journal.Balance{{Number: -123, Unit: u}},
					"Expenses:Food": journal.Balance{{Number: 123, Unit: u}},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  journal.Balance{{Number: -232, Unit: u}},
				Actual:    journal.Balance{{Number: -123, Unit: u}},
				Diff:      journal.Balance{{Number: 109, Unit: u}},
			},
		}
		if diff := cmp.Diff(want, got.Entries); diff != "" {
			t.Errorf("entry mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := TBalance{
			"Assets:Cash":   journal.Balance{{Number: -123, Unit: u}},
			"Expenses:Food": journal.Balance{{Number: 123, Unit: u}},
		}
		compareBalances(t, want, got.Balance)
	})
}

func compareBalances(t *testing.T, want, got TBalance) {
	t.Helper()
	wantKeys := make(map[journal.Account]struct{})
	for k := range want {
		wantKeys[k] = struct{}{}
	}
	for k := range got {
		if !got[k].Equal(want[k]) {
			t.Errorf("For %s got balance %s, want %s", k, got[k], want[k])
		}
		delete(wantKeys, k)
	}
	for k := range wantKeys {
		t.Errorf("For %s missing, want balance %s", k, want[k])
	}
}
