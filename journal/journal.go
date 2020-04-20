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

// Package journal implements bookkeeping with keeper files.
package journal

import (
	"fmt"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/kpr/scanner"
)

// A Book represents accounting information compiled from keeper file source.
type Book struct {
	// Entries are all of the entries, sorted chronologically.
	Entries []Entry
	// AccountEntries are the entries that affect each account.
	AccountEntries map[Account][]Entry
	// Balances is the final balance for all accounts.
	Balances TBalance
}

// BalanceErr returns non-nil if the book has balance assertion errors.
func (b *Book) BalanceErr() error {
	var err scanner.ErrorList
	for _, e := range b.Entries {
		switch e := e.(type) {
		case BalanceAssert:
			if !e.Diff.Empty() {
				msg := fmt.Sprintf("balance for %s declared to be %s, but was %s (diff %s)",
					e.Account, e.Declared, e.Actual, e.Diff)
				err.Add(e.EntryPos, msg)
			}
		}
	}
	return err.Err()
}

// An Option is passed to Compile to configure compilation.
type Option interface {
	option()
}

// Compile compiles keeper file source into a Book.
// Balance assertion errors are not returned here, to enable the
// caller to inspect the transactions to identify the error.
func Compile(src []byte, o ...Option) (*Book, error) {
	e, err := buildEntries(src)
	if err != nil {
		return nil, err
	}
	sortEntries(e)
	op := buildOptions(o)
	if d := op.ending; d.IsValid() {
		e = entriesEnding(e, d)
	}
	b := compile(e)
	return b, nil
}

// Ending returns an option that limits a compiled book to entries
// ending on the given date.
func Ending(d civil.Date) Option {
	return optionSetter(func(o *options) {
		o.ending = d
	})
}

func buildOptions(o []Option) options {
	var op options
	for _, o := range o {
		o.(optionSetter)(&op)
	}
	return op
}

type optionSetter func(*options)

func (optionSetter) option() {}

type options struct {
	ending civil.Date
}

// compile compiles a Book from entries.
// Entries should be sorted.
func compile(e []Entry) *Book {
	b := &Book{
		AccountEntries: make(map[Account][]Entry),
		Balances:       make(TBalance),
	}
	for _, e := range e {
		b.compileEntry(e)
	}
	return b
}

func (b *Book) compileEntry(e Entry) {
	switch e := e.(type) {
	case Transaction:
		e.Balances = make(TBalance)
		for _, s := range e.Splits {
			b.Balances.Add(s.Account, s.Amount)
			e.Balances[s.Account] = b.Balances[s.Account].Copy()
		}
		b.addEntry(e)
	case BalanceAssert:
		bal, ok := b.Balances[e.Account]
		switch ok {
		case true:
			bal = bal.Copy()
		case false:
			bal = make(Balance)
		}
		e.Actual = bal
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
	diff := x.Copy()
	for _, a := range y.Amounts() {
		diff.Sub(a)
	}
	return diff
}
