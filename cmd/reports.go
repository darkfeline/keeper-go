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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/cmd/internal/colfmt"
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
		wf, err := writeBalancesFunc(format)
		if err != nil {
			return err
		}
		wf(os.Stdout, m, "Assets")
		fmt.Println()
		wf(os.Stdout, m, "Liabilities")
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
		wf, err := writeBalancesFunc(format)
		if err != nil {
			return err
		}
		wf(os.Stdout, m, "Income")
		fmt.Println()
		wf(os.Stdout, m, "Expenses")
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

type writeBalancesFn func(w io.Writer, m map[book.Account]book.Balance, root book.Account) error

func writeBalancesFunc(format string) (writeBalancesFn, error) {
	switch format {
	case tabFmt:
		return writeBalancesTab, nil
	case prettyFmt:
		return writeBalancesPretty, nil
	default:
		return nil, fmt.Errorf("unknown format %v", format)
	}
}

func writeBalancesTab(w io.Writer, m map[book.Account]book.Balance, root book.Account) error {
	as := accountsUnder(m, root)
	bw := bufio.NewWriter(w)
	var total book.Balance
	for _, a := range as {
		bw.WriteString(string(a))
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
	}
	bw.WriteString("Total")
	bw.WriteByte('\t')
	bw.WriteString(total.String())
	bw.WriteByte('\n')
	return bw.Flush()
}

// writeBalancesPretty writes balances prettily.
// The amounts are right justified and aligned.
// The units are left justified and aligned.
// If there is more than one unit type in an account,
// its balance is printed comma separated, aligned after the units for
// single unit accounts.
// These are assumed to be trading accounts and less important.
func writeBalancesPretty(w io.Writer, m map[book.Account]book.Balance, root book.Account) error {
	items := makeBalanceItems(m, root)
	return colfmt.Format(w, items)
}

// balanceItem is used to prepare balances for pretty formatting.
type balanceItem struct {
	prefix       string
	amount       string `colfmt:"right"`
	unit         string
	extraBalance string
}

func (i *balanceItem) addBalance(b book.Balance) {
	switch len(b) {
	case 0:
	case 1:
		a := b[0]
		i.amount = a.Scalar()
		i.unit = a.UnitType.Symbol
	default:
		i.extraBalance = b.String()
	}
}

func makeBalanceItems(m map[book.Account]book.Balance, root book.Account) []balanceItem {
	var items []balanceItem
	var total book.Balance
	rlen := len(root.Parts())
	_ = book.WalkAccountTree(accountsUnder(m, root), func(n book.AccountNode) error {
		a := n.Account
		if !a.Under(root) && a != root {
			return nil
		}
		i := balanceItem{
			prefix: indent(len(a.Parts())-rlen) + a.Leaf(),
		}
		b := m[a]
		i.addBalance(b)
		items = append(items, i)
		for _, a := range b {
			total = total.Add(a)
		}
		return nil
	})
	i := balanceItem{
		prefix: "Total",
	}
	i.addBalance(total)
	items = append(items, i)
	return items
}

func indent(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("  ")
	}
	return b.String()
}
