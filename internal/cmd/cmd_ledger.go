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
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/internal/colfmt"
	"go.felesatra.moe/keeper/journal"
)

func init() {
	rootCmd.AddCommand(ledgerCmd)
}

var ledgerCmd = &cobra.Command{
	Use:   "ledger [file] [account]",
	Short: "Print ledger for account",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := compileFile(args[0])
		if err != nil {
			return err
		}
		a := journal.Account(args[1])
		items := makeLedgerItems(a, b.AccountEntries[a])
		f, err := getFormatter(format)
		if err != nil {
			return err
		}
		_ = f(os.Stdout, items)
		return nil
	},
}

type ledgerItem struct {
	date        string
	line        string
	description string
	amount      string `colfmt:"right"`
	balance     string `colfmt:"right"`
	balance2    string `colfmt:"right"`
	balancex    string
	error       string
}

func (l *ledgerItem) setBalance(b journal.Balance) {
	switch len(b) {
	default:
		l.balancex = "(more)"
		fallthrough
	case 2:
		l.balance2 = b[1].String()
		fallthrough
	case 1:
		l.balance = b[0].String()
	case 0:
	}
}

func makeLedgerItems(a journal.Account, e []journal.Entry) []ledgerItem {
	var items []ledgerItem
	for _, e := range e {
		i := ledgerItem{
			date: e.Date().String(),
			line: fmt.Sprintf("L%d", e.Position().Line),
		}

		switch e := e.(type) {
		case journal.Transaction:
			i := i
			i.description = e.Description
			if len(e.Splits) == 0 {
				panic(fmt.Sprintf("no splits for %#v", e))
			}
			for _, s := range e.Splits {
				if s.Account != a {
					continue
				}
				i := i
				i.amount = s.Amount.String()
				items = append(items, i)
			}
			items[len(items)-1].setBalance(e.Balances[a])
		case journal.BalanceAssert:
			if e.Account != a {
				panic(fmt.Sprintf("got balance for account %s not %s", e.Account, a))
			}
			i := i
			i.description = "(balance)"
			i.setBalance(e.Actual)
			if !e.Diff.Empty() {
				i.error = fmt.Sprintf("declared %s (diff %s)", e.Declared, e.Diff)
			}
			items = append(items, i)
		default:
			panic(fmt.Sprintf("unknown entry type %T", e))
		}
	}
	return items
}

type formatter func(io.Writer, interface{}) error

func getFormatter(format string) (formatter, error) {
	switch format {
	case tabFmt:
		return colfmt.FormatTab, nil
	case prettyFmt:
		return colfmt.Format, nil
	default:
		return nil, fmt.Errorf("unknown format %v", format)
	}
}
