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
	"os"

	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse"
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
		a := book.Account(args[1])
		items := makeLedgerItems(a, b.AccountEntries)
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

func (l *ledgerItem) setBalance(b book.Balance) {
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

func makeLedgerItems(a book.Account, l []interface{}) []ledgerItem {
	var is []ledgerItem
	var b book.Balance
lines:
	for _, l := range l {
		switch l := l.(type) {
		case parse.TransactionLine:
		splits:
			for _, s := range l.Splits {
				if s.Account != a {
					continue splits
				}
				li := ledgerItem{
					date:        l.Date.String(),
					line:        fmt.Sprintf("L%d", l.Line),
					description: l.Description,
					amount:      s.Amount.String(),
				}
				if err := l.Err; err != nil {
					li.error = err.Error()
				}
				b = b.Add(s.Amount)
				li.setBalance(b)
				is = append(is, li)
			}
		case parse.BalanceLine:
			if l.Account != a {
				continue lines
			}
			li := ledgerItem{
				date: l.Date.String(),
				line: fmt.Sprintf("L%d", l.Line),
			}
			if err := l.Err; err != nil {
				li.error = err.Error()
			}
			li.setBalance(b)
			is = append(is, li)
		default:
			panic(fmt.Sprintf("unknown line type %T", l))
		}
	}
	return is
}
