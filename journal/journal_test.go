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
			},
			&BalanceAssert{
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
			"Assets:Cash":    Balance{u: -246},
			"Expenses:Food":  Balance{u: 123},
			"Expenses:Drink": Balance{u: 123},
		}
		compareBalances(t, want, got.Balances)
	})
}

func TestCompile_treebal(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "initial",
			Splits: []Split{
				split("Equity:Capital", 540, u),
				split("Assets:Foo", 120, u),
				split("Assets:Bar", 150, u),
				split("Assets:Bar:Eriko", 130, u),
				split("Assets:Bar:Shizuru", 140, u),
			},
		},
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Bar",
			Tree:      true,
			Declared:  Balance{u: 410},
		},
	}
	got, err := compile(e)
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 2},
			Description: "initial",
			Splits: []Split{
				split("Equity:Capital", 540, u),
				split("Assets:Foo", 120, u),
				split("Assets:Bar", 150, u),
				split("Assets:Bar:Eriko", 130, u),
				split("Assets:Bar:Shizuru", 140, u),
			},
		},
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Bar",
			Tree:      true,
			Declared:  Balance{u: 410},
			Actual:    Balance{u: 420},
			Diff:      Balance{u: 10},
		},
	}
	if diff := cmp.Diff(want, got.Entries); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestCompile_tx_after_disable(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		&DisableAccount{
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

func TestCompile_disable_nonempty_account(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		&Transaction{
			EntryDate:   civil.Date{2000, 1, 1},
			Description: "buy stuff",
			Splits: []Split{
				split("Income:Salary", -123, u),
				split("Assets:Cash", 123, u),
			},
		},
		&DisableAccount{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
		},
	}
	got, err := compile(e)
	if err != nil {
		t.Fatal(err)
	}
	want := []*BalanceAssert{
		{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Actual:    Balance{u: 123},
			Diff:      Balance{u: 123},
		},
	}
	if diff := cmp.Diff(got.BalanceErrors, want); diff != "" {
		t.Errorf("balance errors mismatch (-want +got):\n%s", diff)
	}
}

func TestTreeBalance(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	b := Balances{
		"Assets:Foo":         {u: 120},
		"Assets:Bar":         {u: 150},
		"Assets:Bar:Eriko":   {u: 130},
		"Assets:Bar:Shizuru": {u: 140},
	}
	got := treeBalance(b, "Assets:Bar")
	want := Balance{u: 420}
	if !got.Equal(want) {
		t.Errorf("treeBalance() = %s; want %s", got, want)
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
