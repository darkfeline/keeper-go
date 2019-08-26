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

package stage1

import (
	"bufio"
	"io"

	"go.felesatra.moe/keeper"
)

func WriteBalanceSheet(w io.Writer, b Balances) error {
	bw := bufio.NewWriter(w)
	writeAccountTree(bw, b, "Assets")
	writeAccountTree(bw, b, "Liabilities")
	return bw.Flush()
}

func writeAccountTree(bw *bufio.Writer, b Balances, parent keeper.Account) {
	var as []keeper.Account
	for _, a := range b.Accounts() {
		if a.Under(parent) {
			as = append(as, a)
		}
	}

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

func WriteIncomeStatement(w io.Writer, b Balances) error {
	bw := bufio.NewWriter(w)
	writeAccountTree(bw, b, "Revenues")
	writeAccountTree(bw, b, "Expenses")
	return bw.Flush()
}
