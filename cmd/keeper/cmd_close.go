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
	"flag"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/chart"
	"go.felesatra.moe/keeper/internal/month"
	"go.felesatra.moe/keeper/journal"
)

var closeCmd = &command{
	usageLine: "close [-m month] [files]",
	run: func(cmd *command, args []string) {
		fs := flag.NewFlagSet(cmd.name(), flag.ExitOnError)
		m := fs.String("month", "", "Month to close")
		if err := fs.Parse(args); err != nil {
			panic(err)
		}
		if fs.NArg() < 1 {
			cmd.printUsage()
			os.Exit(2)
		}

		var d civil.Date
		if *m != "" {
			var err error
			d, err = month.Parse(*m)
			if err != nil {
				fmt.Fprintf(os.Stderr, "keeper: %s\n", err)
				os.Exit(2)
			}
		} else {
			d = month.Prev(month.Now())
		}

		o := []journal.Option{
			journal.File(fs.Args()...),
			journal.Ending(month.LastDay(d)),
		}
		j, err := journal.Compile(o...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "keeper: %s\n", err)
			os.Exit(1)
		}
		checkBalanceErrsAndExit(j)
		_ = printClosingTx(os.Stdout, j, month.Next(d))
	},
}

func printClosingTx(w io.Writer, j *journal.Journal, d civil.Date) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "tx %s \"Closing\"\n", d)
	c := chart.New(j.Accounts())
	b := make(journal.Balance)
	for _, a := range c.Income() {
		for _, am := range j.Balances[a].Amounts() {
			fmt.Fprintf(bw, "%s %s\n", a, am.Neg())
			b.Add(am)
		}
	}
	for _, a := range c.Expenses() {
		for _, am := range j.Balances[a].Amounts() {
			fmt.Fprintf(bw, "%s %s\n", a, am.Neg())
			b.Add(am)
		}
	}
	a := c.Equity()[0]
	for _, am := range b.Amounts() {
		fmt.Fprintf(bw, "%s %s\n", a, am)
	}
	fmt.Fprintf(bw, "end\n")
	return bw.Flush()
}
