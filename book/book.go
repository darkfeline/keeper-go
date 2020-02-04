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
	"fmt"

	"go.felesatra.moe/keeper/journal"
)

type Book struct {
	Entries        []Entry
	AccountEntries map[journal.Account][]Entry
	Balance        TBalance
}

// Compile compiles entries into a book.
// The input slice is reordered in place.
func Compile(e []journal.Entry) *Book {
	return CompileWithBalance(e, make(TBalance))
}

// CompileWithBalance compiles entries into a book, with the given initial balance.
// The input slice is reordered in place.
func CompileWithBalance(e []journal.Entry, initial TBalance) *Book {
	journal.SortByDate(e)
	b := &Book{
		AccountEntries: make(map[journal.Account][]Entry),
		Balance:        initial,
	}
	for _, e := range e {
		switch e := e.(type) {
		case journal.Transaction:
			tbal := make(TBalance)
			for _, s := range e.Splits {
				k := s.Account
				b.Balance[k] = b.Balance[k].Add(s.Amount)
				tbal[k] = b.Balance[k]
			}
			r := Transaction{
				EntryPos:    e.EntryPos,
				EntryDate:   e.EntryDate,
				Description: e.Description,
				Splits:      e.Splits,
				Balances:    tbal,
			}
			b.addEntry(r)
		case journal.BalanceAssert:
			r := BalanceAssert{
				EntryPos:  e.EntryPos,
				EntryDate: e.EntryDate,
				Account:   e.Account,
				Declared:  e.Balance,
				Actual:    b.Balance[e.Account],
			}
			r.Diff = balanceDiff(r.Actual, r.Declared)
			b.addEntry(r)
		default:
			panic(fmt.Sprintf("unknown Entry type %T", e))
		}
	}
	return b
}

func (b *Book) addEntry(e Entry) {
	b.Entries = append(b.Entries, e)
	switch e := e.(type) {
	case Transaction:
		seen := make(map[journal.Account]bool)
		for _, s := range e.Splits {
			if seen[s.Account] {
				continue
			}
			b.addAccountEntry(s.Account, e)
			seen[s.Account] = true
		}
	case BalanceAssert:
		b.addAccountEntry(e.Account, e)
	default:
		panic(fmt.Sprintf("unknown Entry type %T", e))
	}
}

func (b *Book) addAccountEntry(a journal.Account, e Entry) {
	m, k := b.AccountEntries, a
	m[k] = append(m[k], e)
}

func balanceDiff(x, y journal.Balance) journal.Balance {
	diff := x.CleanCopy()
	for _, a := range y {
		a.Number = -a.Number
		diff.Add(a)
	}
	return diff.CleanCopy()
}
