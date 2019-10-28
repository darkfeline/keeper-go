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
	"io"
	"sort"

	"go.felesatra.moe/keeper/book"
)

func tallyBalances(ts []book.Transaction) map[book.Account]book.Balance {
	m := make(map[book.Account]book.Balance)
	for _, t := range ts {
		for _, s := range t.Splits {
			b := m[s.Account]
			m[s.Account] = b.Add(s.Amount)
		}
	}
	return m
}

func writeAccountTree(w io.Writer, m map[book.Account]book.Balance, root book.Account) {
	var as []book.Account
	for _, a := range m {
		if a.Under(root) {
			as = append(as, a)
		}
	}
	sortAccounts(as)

	bw.WriteString(string(parent))
	bw.WriteString("\t\t\n")
	var total []keeper.Quantity
	pflen := len(parent.Parts())
	_ = keeper.WalkAccountTree(as, func(n keeper.AccountNode) error {
		a := n.Account
		for _ = range a.Parts()[pflen:] {
			bw.WriteString("    ")
		}
		bw.WriteString(a.Leaf())
		bw.WriteString("\t\t\n")
		for _, q := range b[a] {
			bw.WriteByte('\t')
			bw.WriteString(q.String())
			bw.WriteString("\t\n")
			total = keeper.AddQuantity(total, q)
		}
		return nil
	})
	bw.WriteString("\nTotal")
	for _, q := range total {
		bw.WriteByte('\t')
		bw.WriteString(q.String())
		bw.WriteString("\t\n")
	}
}

func sortAccounts(as []book.Account) {
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
}
