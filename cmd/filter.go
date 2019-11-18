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

package cmd

import (
	"sort"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

// txStarting returns the transactions with dates on or after the given date.
func txStarting(ts []book.Transaction, d civil.Date) []book.Transaction {
	i := sort.Search(len(ts), func(i int) bool {
		return !ts[i].Date.Before(d)
	})
	new := make([]book.Transaction, len(ts)-i)
	copy(new, ts[i:])
	return new
}

// txEnding returns the transactions with dates before or on the given date.
func txEnding(ts []book.Transaction, d civil.Date) []book.Transaction {
	i := sort.Search(len(ts), func(i int) bool {
		return !ts[len(ts)-1-i].Date.After(d)
	})
	n := len(ts) - i
	new := make([]book.Transaction, n)
	copy(new, ts[:n])
	return new
}
