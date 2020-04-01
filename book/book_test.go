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
			Declared:  Balance{{Number: -232, Unit: u}},
		},
	}
	got := compile(e, make(TBalance))
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
				Balances: TBalance{
					"Assets:Cash":   Balance{{Number: -123, Unit: u}},
					"Expenses:Food": Balance{{Number: 123, Unit: u}},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  Balance{{Number: -232, Unit: u}},
				Actual:    Balance{{Number: -123, Unit: u}},
				Diff:      Balance{{Number: 109, Unit: u}},
			},
		}
		if diff := cmp.Diff(want, got.Entries); diff != "" {
			t.Errorf("entry mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := TBalance{
			"Assets:Cash":   Balance{{Number: -123, Unit: u}},
			"Expenses:Food": Balance{{Number: 123, Unit: u}},
		}
		compareBalances(t, want, got.Balance)
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
			Declared:  Balance{{Number: -232, Unit: u}},
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
			Declared:  Balance{{Number: -232, Unit: u}},
		},
	}
	got := compile(e, make(TBalance))
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
				Balances: TBalance{
					"Assets:Cash":   Balance{{Number: -123, Unit: u}},
					"Expenses:Food": Balance{{Number: 123, Unit: u}},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  Balance{{Number: -232, Unit: u}},
				Actual:    Balance{{Number: -123, Unit: u}},
				Diff:      Balance{{Number: 109, Unit: u}},
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
				Balances: TBalance{
					"Assets:Cash":   Balance{{Number: -246, Unit: u}},
					"Expenses:Food": Balance{{Number: 246, Unit: u}},
				},
			},
			BalanceAssert{
				EntryDate: civil.Date{2000, 1, 3},
				Account:   "Assets:Cash",
				Declared:  Balance{{Number: -232, Unit: u}},
				Actual:    Balance{{Number: -246, Unit: u}},
				Diff:      Balance{{Number: -14, Unit: u}},
			},
		}
		if diff := cmp.Diff(want, got.Entries); diff != "" {
			t.Errorf("entry mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := TBalance{
			"Assets:Cash":   Balance{{Number: -246, Unit: u}},
			"Expenses:Food": Balance{{Number: 246, Unit: u}},
		}
		compareBalances(t, want, got.Balance)
	})
}

func TestBalanceDiff(t *testing.T) {
	t.Parallel()
	t.Run("bug", func(t *testing.T) {
		t.Parallel()
		actual := Balance{}
		declared := Balance{
			{Number: -200, Unit: Unit{Symbol: "USD", Scale: 100}},
		}
		got := balanceDiff(actual, declared)
		want := Balance{
			{Number: 200, Unit: Unit{Symbol: "USD", Scale: 100}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("balance mismatch (-want +got):\n%s", diff)
		}
	})
}

func compareBalances(t *testing.T, want, got TBalance) {
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
