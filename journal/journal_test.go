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
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
)

func TestCompile(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "buy stuff",
			Splits: []Split{
				{
					Account: "Assets:Cash",
					Amount:  Amount{Number: -123, Unit: u},
				},
				{
					Account: "Expenses:Food",
					Amount:  Amount{Number: 123, Unit: u},
				},
			},
		},
		BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Declared:  Balance{u: -232},
		},
	}
	got := compile(e)
	t.Run("entries", func(t *testing.T) {
		want := []Entry{
			Transaction{
				EntryDate:   civil.Date{2000, 1, 2},
				Description: "buy stuff",
				Splits: []Split{
					{
						Account: "Assets:Cash",
						Amount:  Amount{Number: -123, Unit: u},
					},
					{
						Account: "Expenses:Food",
						Amount:  Amount{Number: 123, Unit: u},
					},
				},
				Balances: Balances{
					"Assets:Cash":   Balance{u: -123},
					"Expenses:Food": Balance{u: 123},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  Balance{u: -232},
				Actual:    Balance{u: -123},
				Diff:      Balance{u: 109},
			},
		}
		if diff := cmp.Diff(want, got.Entries); diff != "" {
			t.Errorf("entry mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := Balances{
			"Assets:Cash":   Balance{u: -123},
			"Expenses:Food": Balance{u: 123},
		}
		compareBalances(t, want, got.Balances)
	})
}

func TestCompile_balances(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "buy stuff",
			Splits: []Split{
				{
					Account: "Assets:Cash",
					Amount:  Amount{Number: -123, Unit: u},
				},
				{
					Account: "Expenses:Food",
					Amount:  Amount{Number: 123, Unit: u},
				},
			},
		},
		BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Declared:  Balance{u: -232},
		},
		Transaction{
			EntryDate:   civil.Date{2000, 1, 3},
			Description: "buy stuff",
			Splits: []Split{
				{
					Account: "Assets:Cash",
					Amount:  Amount{Number: -123, Unit: u},
				},
				{
					Account: "Expenses:Food",
					Amount:  Amount{Number: 123, Unit: u},
				},
			},
		},
		BalanceAssert{
			EntryDate: civil.Date{2000, 1, 3},
			Account:   "Assets:Cash",
			Declared:  Balance{u: -232},
		},
	}
	got := compile(e)
	t.Run("entries", func(t *testing.T) {
		want := []Entry{
			Transaction{
				EntryDate:   civil.Date{2000, 1, 2},
				Description: "buy stuff",
				Splits: []Split{
					{
						Account: "Assets:Cash",
						Amount:  Amount{Number: -123, Unit: u},
					},
					{
						Account: "Expenses:Food",
						Amount:  Amount{Number: 123, Unit: u},
					},
				},
				Balances: Balances{
					"Assets:Cash":   Balance{u: -123},
					"Expenses:Food": Balance{u: 123},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  Balance{u: -232},
				Actual:    Balance{u: -123},
				Diff:      Balance{u: 109},
			},
			Transaction{
				EntryDate:   civil.Date{2000, 1, 3},
				Description: "buy stuff",
				Splits: []Split{
					{
						Account: "Assets:Cash",
						Amount:  Amount{Number: -123, Unit: u},
					},
					{
						Account: "Expenses:Food",
						Amount:  Amount{Number: 123, Unit: u},
					},
				},
				Balances: Balances{
					"Assets:Cash":   Balance{u: -246},
					"Expenses:Food": Balance{u: 246},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 3},
				Account:   "Assets:Cash",
				Declared:  Balance{u: -232},
				Actual:    Balance{u: -246},
				Diff:      Balance{u: -14},
			},
		}
		if diff := cmp.Diff(want, got.Entries); diff != "" {
			t.Errorf("entries mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := Balances{
			"Assets:Cash":   Balance{u: -246},
			"Expenses:Food": Balance{u: 246},
		}
		compareBalances(t, want, got.Balances)
	})
}

func TestBalanceDiff(t *testing.T) {
	t.Parallel()
	t.Run("bug", func(t *testing.T) {
		t.Parallel()
		u := Unit{Symbol: "USD", Scale: 100}
		actual := Balance{}
		declared := Balance{u: -200}
		got := balanceDiff(actual, declared)
		want := Balance{u: 200}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("balance mismatch (-want +got):\n%s", diff)
		}
	})
}

func compareBalances(t *testing.T, want, got Balances) {
	t.Helper()
	wantKeys := make(map[Account]struct{})
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