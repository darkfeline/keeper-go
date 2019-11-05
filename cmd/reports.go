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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse"
	"go.felesatra.moe/keeper/report"
)

func init() {
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(incomeCmd)
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Print balance sheet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		ts, err := parse.Parse(f)
		if err != nil {
			return err
		}
		m := report.TallyBalances(ts)
		writeAccountTree(os.Stdout, m, "Assets")
		writeAccountTree(os.Stdout, m, "Liabilities")
		return nil
	},
}

var incomeCmd = &cobra.Command{
	Use:   "income",
	Short: "Print income statement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		ts, err := parse.Parse(f)
		if err != nil {
			return err
		}
		m := report.TallyBalances(ts)
		writeAccountTree(os.Stdout, m, "Income")
		writeAccountTree(os.Stdout, m, "Expenses")
		return nil
	},
}

func accountsUnder(m map[book.Account]book.Balance, root book.Account) []book.Account {
	var as []book.Account
	for a := range m {
		if a.Under(root) {
			as = append(as, a)
		}
	}
	report.SortAccounts(as)
	return as
}

func writeAccountTree(w io.Writer, m map[book.Account]book.Balance, root book.Account) error {
	as := accountsUnder(m, root)
	bw := bufio.NewWriter(w)
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
			bw.WriteByte('\t')
			bw.WriteByte('\t')
			bw.WriteByte('\n')
			return nil
		}
		bw.WriteByte('\t')
		bw.WriteString(b.String())
		bw.WriteByte('\t')
		bw.WriteByte('\n')
		for _, a := range b {
			total = total.Add(a)
		}
		return nil
	})
	bw.WriteString("Total")
	bw.WriteByte('\t')
	bw.WriteString(total.String())
	bw.WriteByte('\t')
	bw.WriteByte('\n')
	return bw.Flush()
}

func indent(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("  ")
	}
	return b.String()
}
