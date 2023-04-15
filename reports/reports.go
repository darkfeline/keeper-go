// Copyright (C) 2021  Allen Li
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

/*
Package reports implements the production of various common reports
from bookkeeping data.
*/
package reports

import (
	"fmt"
	"sort"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/journal"
)

// A Pair is a Debit-Credit pair.
type Pair[T any] struct {
	Debit  T
	Credit T
}

// A TrialBalance represents a trial balance report.
type TrialBalance struct {
	Rows  []TrialBalanceRow
	Total Pair[journal.Balance]
}

// A TrialBalanceRow represents a row in a TrialBalance.
type TrialBalanceRow struct {
	Account journal.Account
	Pairs   []Pair[*journal.Amount]
}

// NewTrialBalance creates a trial balance report.
func NewTrialBalance(j *journal.Journal) *TrialBalance {
	b := j.Balances
	var total Pair[journal.Balance]
	var r []TrialBalanceRow
	for _, a := range sortedAccounts(j) {
		e := TrialBalanceRow{
			Account: a,
		}
		for _, amt := range b[a].Amounts() {
			p := Pair[*journal.Amount]{}
			switch amt.Number.Sign() {
			case -1:
				p.Credit = amt
				total.Credit.Add(amt)
			case 1:
				p.Debit = amt
				total.Debit.Add(amt)
			}
			e.Pairs = append(e.Pairs, p)
		}
		r = append(r, e)
	}
	return &TrialBalance{
		Rows:  r,
		Total: total,
	}
}

// An AccountLedger represents the ledger for one account.
type AccountLedger struct {
	Account journal.Account
	Rows    []LedgerRow
}

// A LedgerRow represents a row in an AccountLedger.
type LedgerRow struct {
	Date        civil.Date
	Description string
	// A reference to the file location for the transaction split.
	Ref  string
	Pair Pair[*journal.Amount]
	// Running balance for the account.
	Balance journal.Balance
}

func NewAccountLedger(j *journal.Journal, a journal.Account) *AccountLedger {
	l := &AccountLedger{Account: a}
	var b journal.Balance
	for _, e := range j.Entries {
		r := LedgerRow{
			Date: e.Date(),
			Ref:  e.Position().String(),
		}
		switch e := e.(type) {
		case *journal.Transaction:
			r.Description = e.Description
			for _, s := range e.Splits {
				r := r
				if s.Account != a {
					continue
				}
				switch s.Amount.Sign() {
				case -1:
					r.Pair.Credit = s.Amount
				case 1:
					r.Pair.Debit = s.Amount
				}
				b.Add(s.Amount)
				r.Balance.Set(&b)
				l.Rows = append(l.Rows, r)
			}
		case *journal.BalanceAssert:
			if e.Account != a {
				break
			}
			units := allUnits(e.Actual, e.Declared, e.Diff)
			t := "balance"
			if e.Tree {
				t = "tree balance"
			}
			for _, u := range units {
				if !e.Diff.Has(u) {
					r.Description = "(" + t + ")"
				} else {
					r.Description = fmt.Sprintf("(%s error, declared %s, diff %s)",
						t, e.Declared.Amount(u), e.Diff.Amount(u))
				}
				r.Balance.Set(&b)
				l.Rows = append(l.Rows, r)
			}
		case *journal.DisableAccount:
			if e.Account != a {
				break
			}
			r.Description = "(disabled)"
			r.Balance.Set(&b)
			l.Rows = append(l.Rows, r)
		default:
			panic(e)
		}
	}
	return l
}

// A Ledger represents a ledger for one or more accounts.
type Ledger struct {
	Accounts []journal.Account
	Rows     []FreeLedgerRow
}

// A FreeLedgerRow represents a row in a Ledger.
type FreeLedgerRow struct {
	Date        civil.Date
	Description string
	// A reference to the file location for the transaction split.
	Ref     string
	Account journal.Account
	Pair    Pair[*journal.Amount]
	// Running balance for all accounts combined.
	Balance journal.Balance
}

// NewLedger creates a Ledger.
// The account slice is referenced directly in the returned Ledger.
func NewLedger(j *journal.Journal, a ...journal.Account) *Ledger {
	l := &Ledger{Accounts: a}
	accs := make(map[journal.Account]bool)
	for _, a := range a {
		accs[a] = true
	}
	var b journal.Balance
	for _, e := range j.Entries {
		r := FreeLedgerRow{
			Date: e.Date(),
			Ref:  e.Position().String(),
		}
		switch e := e.(type) {
		case *journal.Transaction:
			r.Description = e.Description
			for _, s := range e.Splits {
				r := r
				if !accs[s.Account] {
					continue
				}
				r.Account = s.Account
				switch s.Amount.Sign() {
				case -1:
					r.Pair.Credit = s.Amount
				case 1:
					r.Pair.Debit = s.Amount
				}
				b.Add(s.Amount)
				r.Balance.Set(&b)
				l.Rows = append(l.Rows, r)
			}
		case *journal.BalanceAssert:
			// XXXXXXXX how to handle treebal?
			if !accs[e.Account] {
				break
			}
			units := allUnits(e.Actual, e.Declared, e.Diff)
			t := "balance"
			if e.Tree {
				t = "tree balance"
			}
			for _, u := range units {
				if !e.Diff.Has(u) {
					r.Description = "(" + t + ")"
				} else {
					r.Description = fmt.Sprintf("(%s error, declared %s, diff %s)",
						t, e.Declared.Amount(u), e.Diff.Amount(u))
				}
				r.Balance.Set(&b)
				l.Rows = append(l.Rows, r)
			}
		case *journal.DisableAccount:
			if !accs[e.Account] {
				break
			}
			r.Description = "(disabled)"
			r.Account = e.Account
			r.Balance.Set(&b)
			l.Rows = append(l.Rows, r)
		default:
			panic(e)
		}
	}
	return l
}

// allUnits returns all of the units in the balances.
func allUnits(b ...journal.Balance) []journal.Unit {
	seen := make(map[journal.Unit]bool)
	for _, b := range b {
		for _, u := range b.Units() {
			seen[u] = true
		}
	}
	var units []journal.Unit
	for u, v := range seen {
		if v {
			units = append(units, u)
		}
	}
	sort.Slice(units, func(i, j int) bool { return units[i].Symbol < units[j].Symbol })
	return units
}

func sortedAccounts(j *journal.Journal) []journal.Account {
	var new []journal.Account
	for a := range j.Accounts {
		new = append(new, a)
	}
	sortAccounts(new)
	return new
}

func sortAccounts(a []journal.Account) {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
}
