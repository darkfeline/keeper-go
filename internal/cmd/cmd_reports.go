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
	"go.felesatra.moe/keeper/internal/colfmt"
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
		b, err := compileFile(args[0])
		if err != nil {
			return err
		}
		f, err := getTbalFormatter()
		if err != nil {
			return err
		}
		f(os.Stdout, b.Balance, "Assets")
		fmt.Println()
		f(os.Stdout, b.Balance, "Liabilities")
		return nil
	},
}

var incomeCmd = &cobra.Command{
	Use:   "income",
	Short: "Print income statement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := compileFile(args[0])
		if err != nil {
			return err
		}
		f, err := getTbalFormatter()
		if err != nil {
			return err
		}
		f(os.Stdout, b.Balance, "Income")
		fmt.Println()
		f(os.Stdout, b.Balance, "Expenses")
		return nil
	},
}

type tbalFormatter func(w io.Writer, m book.TBalance, root book.Account) error

func getTbalFormatter() (tbalFormatter, error) {
	switch format {
	case tabFmt:
		return formatTbalTab, nil
	case prettyFmt:
		return formatTbalPretty, nil
	default:
		return nil, fmt.Errorf("unknown format %v", format)
	}
}

func formatTbalTab(w io.Writer, m book.TBalance, root book.Account) error {
	type item struct {
		account string
		balance string
	}
	var is []item
	var total book.Balance
	for _, a := range accountsUnder(m, root) {
		is = append(is, item{
			account: string(a),
			balance: m[a].String(),
		})
		for _, a := range m[a] {
			total = total.Add(a)
		}
	}
	is = append(is, item{
		account: "Total",
		balance: total.String(),
	})

	bw := bufio.NewWriter(w)
	colfmt.FormatTab(bw, is)
	return bw.Flush()
}

// formatTbalPretty writes balances prettily.
// The amounts are right justified and aligned.
// The units are left justified and aligned.
// If there is more than one unit type in an account,
// its balance is printed comma separated, aligned after the units for
// single unit accounts.
// These are assumed to be trading accounts and less important.
func formatTbalPretty(w io.Writer, m book.TBalance, root book.Account) error {
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
		i.unit = a.Unit.Symbol
	default:
		i.extraBalance = b.String()
	}
}

func makeBalanceItems(m book.TBalance, root book.Account) []balanceItem {
	var items []balanceItem
	var total book.Balance
	rlen := root.Level()
	_ = walkAccountTree(accountsUnder(m, root), func(n accountNode) error {
		a := n.Account
		if !a.Under(root) && a != root {
			return nil
		}
		i := balanceItem{
			prefix: indent(a.Level()-rlen) + a.Leaf(),
		}
		b := m[a]
		if n.Leaf && b.Empty() {
			return nil
		}
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

func accountsUnder(m book.TBalance, root book.Account) []book.Account {
	var as []book.Account
	for a := range m {
		if a.Under(root) {
			as = append(as, a)
		}
	}
	sortAccounts(as)
	return as
}

// indent returns a whitespace prefix for the given number of levels
// of indentation.
func indent(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("  ")
	}
	return b.String()
}
