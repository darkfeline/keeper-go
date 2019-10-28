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
	"bufio"
	"io"
	"sort"
	"strings"

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

func writeAccountTree(w io.Writer, m map[book.Account]book.Balance, root book.Account) error {
	bw := bufio.NewWriter(w)
	var as []book.Account
	for a := range m {
		if a.Under(root) {
			as = append(as, a)
		}
	}
	sortAccounts(as)

	bw.WriteString(string(root))
	bw.WriteByte('\n')
	var total book.Balance
	rlen := len(root.Parts())
	_ = book.WalkAccountTree(as, func(n book.AccountNode) error {
		a := n.Account
		if !a.Under(root) && a != root {
			return nil
		}
		prefix := indent(len(a.Parts()) - rlen)
		bw.WriteString(prefix)
		bw.WriteString(a.Leaf())
		b := m[a]
		if len(b) == 0 {
			bw.WriteByte('\n')
			return nil
		}
		bw.WriteByte('\t')
		bw.WriteString(b.String())
		bw.WriteByte('\n')
		for _, a := range b {
			total = total.Add(a)
		}
		return nil
	})
	bw.WriteString("Total")
	bw.WriteByte('\t')
	bw.WriteString(total.String())
	bw.WriteByte('\n')
	return bw.Flush()
}

func indent(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte('\t')
	}
	return b.String()
}

func sortAccounts(as []book.Account) {
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
}
