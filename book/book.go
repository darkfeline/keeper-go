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
)

type Book struct {
	Entries        []Entry
	AccountEntries map[Account][]Entry
	Balance        TBalance
}

func Compile(src []byte) (*Book, error) {
	e, err := buildEntries(src)
	if err != nil {
		return nil, err
	}
	return compileFromEntries(e), nil
}

func compileFromEntries(e []Entry) *Book {
	sortEntries(e)
	b := &Book{
		AccountEntries: make(map[Account][]Entry),
		Balance:        make(TBalance),
	}
	for _, e := range e {
		b.compileEntry(e)
	}
	return b
}

func (b *Book) compileEntry(e Entry) {
	switch e := e.(type) {
	case Transaction:
		tbal := make(TBalance)
		for _, s := range e.Splits {
			k := s.Account
			b.Balance[k] = b.Balance[k].Add(s.Amount)
			tbal[k] = b.Balance[k]
		}
		e.Balances = tbal
		b.addEntry(e)
	case BalanceAssert:
		e.Actual = b.Balance[e.Account]
		e.Diff = balanceDiff(e.Actual, e.Declared)
		b.addEntry(e)
	default:
		panic(fmt.Sprintf("unknown Entry type %T", e))
	}
}

func (b *Book) addEntry(e Entry) {
	b.Entries = append(b.Entries, e)
	switch e := e.(type) {
	case Transaction:
		seen := make(map[Account]bool)
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

func (b *Book) addAccountEntry(a Account, e Entry) {
	m, k := b.AccountEntries, a
	m[k] = append(m[k], e)
}

func balanceDiff(x, y Balance) Balance {
	diff := x.CleanCopy()
	for _, a := range y {
		a.Number = -a.Number
		diff.Add(a)
	}
	return diff.CleanCopy()
}
