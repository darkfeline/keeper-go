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

package cmd

import (
	"sort"
	"strings"

	"go.felesatra.moe/keeper/book"
)

func acctConcat(parts []string) book.Account {
	return book.Account(strings.Join(":", parts))
}

func acctParts(a book.Account) []string {
	return strings.Split(string(a), ":")
}

// acctLeaf returns the leaf part of the Account (after the last colon).
func acctLeaf(a book.Account) string {
	p := a.Parts()
	return p[len(p)-1]
}

// acctUnder returns true if the Account is a child account of the argument.
func acctUnder(a book.Account, parent book.Account) bool {
	return strings.HasPrefix(string(a), string(parent)+":")
}

func sortAccounts(as []book.Account) {
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
}
