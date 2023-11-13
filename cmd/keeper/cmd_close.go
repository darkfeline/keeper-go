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

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/internal/month"
	"go.felesatra.moe/keeper/journal"
)

var closeCmd = &command{
	usageLine: "close [-month month] [-trading] [files]",
	run: func(cmd *command, args []string) {
		fs := cmd.flagSet()
		c := configPath(fs)
		m := fs.String("month", "", "Month to close")
		t := fs.Bool("trading", false, "Include trading accounts")
		fs.Parse(args)
		if fs.NArg() < 1 {
			fs.Usage()
			os.Exit(2)
		}

		var d civil.Date
		if *m != "" {
			var err error
			d, err = month.Parse(*m)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			d = month.Prev(month.Now())
		}

		j, err := journal.Compile(&journal.CompileArgs{
			Inputs: journal.Files(fs.Args()...),
			Ending: month.LastDay(d),
		})
		if err != nil {
			log.Fatal(err)
		}
		checkBalanceErrsAndExit(j)
		var accsToClose []journal.Account
		var equity journal.Account
		for ac := range j.Accounts {
			switch {
			case c.IsIncome(ac), c.IsExpenses(ac):
				accsToClose = append(accsToClose, ac)
			case *t && c.IsTrading(ac):
				accsToClose = append(accsToClose, ac)
			case equity == "" && c.IsEquity(ac):
				equity = ac
			}
		}
		sort.Slice(accsToClose, func(i, j int) bool { return accsToClose[i] < accsToClose[j] })
		_ = printClosingTx(os.Stdout, j, d, equity, accsToClose)
		_ = printClosingBalances(os.Stdout, month.Next(d), accsToClose)
	},
}

func printClosingBalances(w io.Writer, d civil.Date, a []journal.Account) error {
	bw := bufio.NewWriter(w)
	for _, a := range a {
		fmt.Fprintf(bw, "balance %s %s 0 USD\n", d, a)
	}
	return bw.Flush()
}

// printClosingTx prints a transaction entry that moves everything
// from the given accounts (usually income, etc. accounts) into the
// destination account (usually equity account).
func printClosingTx(w io.Writer, j *journal.Journal, d civil.Date, dst journal.Account, a []journal.Account) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "tx %s \"Closing\"\n", d)
	var b journal.Balance
	for _, a := range a {
		for _, am := range j.Balances[a].Amounts() {
			am.Neg()
			fmt.Fprintf(bw, "%s %s\n", a, am)
			b.Add(am)
		}
	}
	b.Neg()
	for _, am := range b.Amounts() {
		fmt.Fprintf(bw, "%s %s\n", dst, am)
	}
	fmt.Fprintf(bw, "end\n")
	return bw.Flush()
}

func filter(a []journal.Account, f func(a journal.Account) bool) []journal.Account {
	var new []journal.Account
	for _, a := range a {
		if f(a) {
			new = append(new, a)
		}
	}
	return new
}
