// Copyright (C) 2019  Allen Li
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

// Package report implements convenient functions for processing
// transactions and making reports.
//
// All functions assume that transactions are sorted and balanced.
package report

import (
	"sort"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

func TallyBalances(ts []book.Transaction) map[book.Account]book.Balance {
	m := make(map[book.Account]book.Balance)
	for _, t := range ts {
		for _, s := range t.Splits {
			b := m[s.Account]
			m[s.Account] = b.Add(s.Amount)
		}
	}
	return m
}

func SortAccounts(as []book.Account) {
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
}

func AccountTx(ts []book.Transaction, a book.Account) []book.Transaction {
	var new []book.Transaction
	for _, t := range ts {
	searchSplits:
		for _, s := range t.Splits {
			if s.Account != a {
				continue
			}
			new = append(new, t)
			break searchSplits
		}
	}
	return new
}

func AccountSplits(ts []book.Transaction, a book.Account) []book.Transaction {
	var new []book.Transaction
	for _, t := range ts {
	searchSplits:
		for _, s := range t.Splits {
			if s.Account != a {
				continue
			}
			t.Splits = []book.Split{s}
			new = append(new, t)
			break searchSplits
		}
	}
	return new
}

func TxStarting(ts []book.Transaction, d civil.Date) []book.Transaction {
	i := sort.Search(len(ts), func(i int) bool {
		return !ts[i].Date.Before(d)
	})
	new := make([]book.Transaction, len(ts)-i)
	copy(new, ts[i:])
	return new
}

func TxEnding(ts []book.Transaction, d civil.Date) []book.Transaction {
	i := sort.Search(len(ts), func(i int) bool {
		return !ts[len(ts)-1-i].Date.After(d)
	})
	n := len(ts) - i
	new := make([]book.Transaction, n)
	copy(new, ts[:n])
	return new
}
