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
	"sort"

	"go.felesatra.moe/keeper/position"
)

func WriteBalanceSheet(w io.Writer, b Balance) error {
	bw := bufio.NewWriter(w)
	bw.WriteString("Assets\n")
	writeAccounts(bw, b[Assets])
	bw.WriteString("\nLiabilities\n")
	writeAccounts(bw, b[Liabilities])
	return bw.Flush()
}

func writeAccounts(bw *bufio.Writer, m TypeBalance) {
	var total []position.Position
	for _, k := range getAccounts(m) {
		bw.WriteString(k)
		for _, p := range m[k] {
			total = position.Add(total, p.Amount, p.Unit)
			bw.WriteByte('\t')
			bw.WriteString(p.String())
			bw.WriteByte('\t')
			bw.WriteByte('\n')
		}
	}
	bw.WriteByte('\n')
	bw.WriteString("Total")
	for _, p := range total {
		bw.WriteByte('\t')
		bw.WriteString(p.String())
		bw.WriteByte('\t')
		bw.WriteByte('\n')
	}
	bw.WriteByte('\n')
}

func getAccounts(m TypeBalance) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func WriteIncomeStatement(w io.Writer, b Balance) error {
	bw := bufio.NewWriter(w)
	bw.WriteString("Revenues\n")
	writeAccounts(bw, b[Revenues])
	bw.WriteString("\nExpenses\n")
	writeAccounts(bw, b[Expenses])
	return bw.Flush()
}
