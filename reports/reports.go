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
	"sort"

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
