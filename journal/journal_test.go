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
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "buy stuff",
			Splits: []Split{
				split("Assets:Cash", -123, u),
				split("Expenses:Food", 123, u),
			},
		},
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Declared:  Balance{u: -232},
		},
	}
	got, err := compile(e)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("entries", func(t *testing.T) {
		want := []Entry{
			&Transaction{
				EntryDate:   civil.Date{2000, 1, 2},
				Description: "buy stuff",
				Splits: []Split{
					split("Assets:Cash", -123, u),
					split("Expenses:Food", 123, u),
				},
				Balances: Balances{
					"Assets:Cash":   Balance{u: -123},
					"Expenses:Food": Balance{u: 123},
				},
			},
			&BalanceAssert{
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
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "buy stuff",
			Splits: []Split{
				split("Assets:Cash", -123, u),
				split("Expenses:Food", 123, u),
			},
		},
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Declared:  Balance{u: -232},
		},
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 3},
			Description: "buy stuff",
			Splits: []Split{
				split("Assets:Cash", -123, u),
				split("Expenses:Drink", 123, u),
			},
		},
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 3},
			Account:   "Assets:Cash",
			Declared:  Balance{u: -232},
		},
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 3},
			Tree:      true,
			Account:   "Expenses",
			Declared:  Balance{u: 321},
		},
	}
	got, err := compile(e)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("entries", func(t *testing.T) {
		want := []Entry{
			&Transaction{
				EntryDate:   civil.Date{2000, 1, 2},
				Description: "buy stuff",
				Splits: []Split{
					split("Assets:Cash", -123, u),
					split("Expenses:Food", 123, u),
				},
				Balances: Balances{
					"Assets:Cash":   Balance{u: -123},
					"Expenses:Food": Balance{u: 123},
				},
			},
			&BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  Balance{u: -232},
				Actual:    Balance{u: -123},
				Diff:      Balance{u: 109},
			},
			&Transaction{
				EntryDate:   civil.Date{2000, 1, 3},
				Description: "buy stuff",
				Splits: []Split{
					split("Assets:Cash", -123, u),
					split("Expenses:Drink", 123, u),
				},
				Balances: Balances{
					"Assets:Cash":    Balance{u: -246},
					"Expenses:Drink": Balance{u: 123},
				},
			},
			&BalanceAssert{
				EntryDate: civil.Date{2000, 1, 3},
				Account:   "Assets:Cash",
				Declared:  Balance{u: -232},
				Actual:    Balance{u: -246},
				Diff:      Balance{u: -14},
			},
			&BalanceAssert{
				EntryDate: civil.Date{2000, 1, 3},
				Account:   "Expenses",
				Tree:      true,
				Declared:  Balance{u: 321},
				Actual:    Balance{u: 246},
				Diff:      Balance{u: -75},
			},
		}
		if diff := cmp.Diff(want, got.Entries); diff != "" {
			t.Errorf("entries mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := Balances{
			"Assets:Cash":    Balance{u: -246},
			"Expenses:Food":  Balance{u: 123},
			"Expenses:Drink": Balance{u: 123},
		}
		compareBalances(t, want, got.Balances)
	})
	t.Run("summary", func(t *testing.T) {
		want := Summary{
			"Assets":         Balance{u: -246},
			"Assets:Cash":    Balance{u: -246},
			"Expenses":       Balance{u: 246},
			"Expenses:Food":  Balance{u: 123},
			"Expenses:Drink": Balance{u: 123},
		}
		compareBalances(t, want, got.Summary)
	})
}

func TestCompile_tx_after_close(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		&CloseAccount{
			EntryDate: civil.Date{2000, 1, 1},
			Account:   "Assets:Cash",
		},
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "buy stuff",
			Splits: []Split{
				split("Assets:Cash", -123, u),
				split("Expenses:Food", 123, u),
			},
		},
	}
	_, err := compile(e)
	if err == nil {
		t.Error("Expected error")
	}
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

func compareBalances(t *testing.T, want, got map[Account]Balance) {
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

func split(a Account, n int64, u Unit) Split {
	return Split{
		Account: a,
		Amount:  Amount{Number: n, Unit: u},
	}
}
