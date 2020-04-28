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
	"sort"
	"time"

	"cloud.google.com/go/civil"

	"go.felesatra.moe/keeper/kpr/token"
)

// All entry types implement Entry.
type Entry interface {
	Date() civil.Date
	Position() token.Position
	sortKey() int64
	entry()
}

// A BalanceAssert entry represents a balance assertion.
type BalanceAssert struct {
	EntryPos  token.Position
	EntryDate civil.Date
	Account   Account
	// Whether this is a tree balance assertion.
	Tree     bool
	Declared Balance
	Actual   Balance
	Diff     Balance // Actual - Declared
}

func (b *BalanceAssert) Position() token.Position {
	return b.EntryPos
}

func (b *BalanceAssert) Date() civil.Date {
	return b.EntryDate
}

func (b *BalanceAssert) sortKey() int64 {
	return dateKey(b.EntryDate) + 1
}

func (*BalanceAssert) entry() {}

// A Transaction entry describes a bookkeeping transaction.
// The total balance of all splits should be zero.
type Transaction struct {
	EntryPos    token.Position
	EntryDate   civil.Date
	Description string
	Splits      []Split
	// Balances contains the balance for all accounts mentioned in
	// the transaction immediately after the transaction.
	Balances Balances
}

func (t *Transaction) Position() token.Position {
	return t.EntryPos
}

func (t *Transaction) Date() civil.Date {
	return t.EntryDate
}

func (t *Transaction) sortKey() int64 {
	return dateKey(t.EntryDate)
}

func (*Transaction) entry() {}

// Split is one split in a transaction.
// This describes a change in the amount of one unit for one account.
type Split struct {
	Account Account
	Amount  Amount
}

// A CloseAccount entry represents an account closing.
type CloseAccount struct {
	EntryPos  token.Position
	EntryDate civil.Date
	Account   Account
}

func (c *CloseAccount) Position() token.Position {
	return c.EntryPos
}

func (c *CloseAccount) Date() civil.Date {
	return c.EntryDate
}

func (c *CloseAccount) sortKey() int64 {
	return dateKey(c.EntryDate) + 2
}

func (*CloseAccount) entry() {}

func sortEntries(e []Entry) {
	type pair struct {
		k int64
		v Entry
	}
	ks := make([]pair, len(e))
	for i, e := range e {
		ks[i] = pair{e.sortKey(), e}
	}
	sort.Slice(ks, func(i, j int) bool {
		return ks[i].k < ks[j].k
	})
	for i, k := range ks {
		e[i] = k.v
	}
}

func dateKey(d civil.Date) int64 {
	return d.In(time.UTC).Unix()
}
