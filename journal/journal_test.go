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
			Declared:  new(balFac).add(u, -232).bal(),
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
				Declared:  new(balFac).add(u, -232).bal(),
				Actual:    new(balFac).add(u, -123).bal(),
				Diff:      new(balFac).add(u, 109).bal(),
			},
		}
		if diff := cmpdiff(want, got.Entries); diff != "" {
			t.Errorf("entry mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := Balances{
			"Assets:Cash":   new(balFac).add(u, -123).pbal(),
			"Expenses:Food": new(balFac).add(u, 123).pbal(),
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
			Declared:  new(balFac).add(u, -232).bal(),
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
			Declared:  new(balFac).add(u, -232).bal(),
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
				Declared:  new(balFac).add(u, -232).bal(),
				Actual:    new(balFac).add(u, -123).bal(),
				Diff:      new(balFac).add(u, 109).bal(),
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
				Declared:  new(balFac).add(u, -232).bal(),
				Actual:    new(balFac).add(u, -246).bal(),
				Diff:      new(balFac).add(u, -14).bal(),
			},
		}
		if diff := cmpdiff(want, got.Entries); diff != "" {
			t.Errorf("entries mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := Balances{
			"Assets:Cash":    new(balFac).add(u, -246).pbal(),
			"Expenses:Food":  new(balFac).add(u, 123).pbal(),
			"Expenses:Drink": new(balFac).add(u, 123).pbal(),
		}
		compareBalances(t, want, got.Balances)
	})
}

func TestCompile_balance_assert_new_account(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	e := []Entry{
		&BalanceAssert{
			EntryDate: civil.Date{2000, 1, 2},
			Account:   "Assets:Cash",
			Declared:  new(balFac).add(u, 0).bal(),
		},
	}
	got, err := compile(e)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("entries", func(t *testing.T) {
		want := []Entry{
			&BalanceAssert{
				EntryDate: civil.Date{2000, 1, 2},
				Account:   "Assets:Cash",
				Declared:  new(balFac).add(u, 0).bal(),
				Actual:    new(balFac).add(u, 0).bal(),
				Diff:      new(balFac).add(u, 0).bal(),
			},
		}
		if diff := cmpdiff(want, got.Entries); diff != "" {
			t.Errorf("entries mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("balance", func(t *testing.T) {
		want := Balances{}
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
			Declared:  new(balFac).add(u, 410).bal(),
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
			Declared:  new(balFac).add(u, 410).bal(),
			Actual:    new(balFac).add(u, 420).bal(),
			Diff:      new(balFac).add(u, 10).bal(),
		},
	}
	if diff := cmpdiff(want, got.Entries); diff != "" {
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
			Actual:    new(balFac).add(u, 123).bal(),
			Diff:      new(balFac).add(u, 123).bal(),
		},
	}
	if diff := cmpdiff(got.BalanceErrors, want); diff != "" {
		t.Errorf("balance errors mismatch (-want +got):\n%s", diff)
	}
}

func TestCompile_account_metadata(t *testing.T) {
	t.Parallel()
	got, err := compileText(`account Some:account
meta "nilou" "nahida"
end`)
	if err != nil {
		t.Fatal(err)
	}
	want := AccountMap{
		"Some:account": &AccountInfo{
			Metadata: map[string]string{"nilou": "nahida"},
		},
	}
	if diff := cmp.Diff(want, got.Accounts); diff != "" {
		t.Errorf("accounts mismatch (-want +got):\n%s", diff)
	}
}

func TestTreeBalance(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	b := Balances{
		"Assets:Foo":         new(balFac).add(u, 120).pbal(),
		"Assets:Bar":         new(balFac).add(u, 150).pbal(),
		"Assets:Bar:Eriko":   new(balFac).add(u, 130).pbal(),
		"Assets:Bar:Shizuru": new(balFac).add(u, 140).pbal(),
	}
	got := new(Balance)
	addTreeBalance(got, b, "Assets:Bar")
	want := new(balFac).add(u, 420).pbal()
	if !got.Equal(want) {
		t.Errorf("treeBalance() = %s; want %s", got, want)
	}
}

func compileText(s string) (*Journal, error) {
	return Compile(&CompileArgs{
		Inputs: []CompileInput{Bytes("testfile", []byte(s))},
	})
}

func compareBalances(t *testing.T, want, got map[Account]*Balance) {
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
	s := Split{Amount: new(Amount), Account: a}
	s.Amount.Number.SetInt64(n)
	s.Amount.Unit = u
	return s
}
