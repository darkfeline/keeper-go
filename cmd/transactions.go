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

	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/cmd/internal/colfmt"
	"go.felesatra.moe/keeper/parse"
	"go.felesatra.moe/keeper/report"
)

func init() {
	rootCmd.AddCommand(splitsCmd)
}

var splitsCmd = &cobra.Command{
	Use:   "splits [file] [account]",
	Short: "Print splits for account",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		r, err := parse.Parse(f)
		if err != nil {
			return err
		}
		ts := report.Transactions(r)
		ts = report.AccountSplits(ts, book.Account(args[1]))
		wf, err := writeSplitsFunc(format)
		if err != nil {
			return err
		}
		wf(os.Stdout, ts)
		return nil
	},
}

type writeSplitsFn func(w io.Writer, ts []book.Transaction) error

func writeSplitsFunc(format string) (writeSplitsFn, error) {
	switch format {
	case tabFmt:
		return writeSplitsTab, nil
	case prettyFmt:
		return writeSplitsPretty, nil
	default:
		return nil, fmt.Errorf("unknown format %v", format)
	}
}

func writeSplitsTab(w io.Writer, ts []book.Transaction) error {
	bw := bufio.NewWriter(w)
	for _, t := range ts {
		for _, s := range t.Splits {
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.Date, t.Description, s.Amount)
		}
	}
	return bw.Flush()
}

func writeSplitsPretty(w io.Writer, ts []book.Transaction) error {
	return colfmt.Format(w, makeSplitItems(ts))
}

type splitItem struct {
	date        string
	description string
	amount      string `colfmt:"right"`
	unit        string
}

func makeSplitItems(ts []book.Transaction) []splitItem {
	var i []splitItem
	for _, t := range ts {
		for _, s := range t.Splits {
			i = append(i, splitItem{
				date:        t.Date.String(),
				description: t.Description,
				amount:      s.Amount.Scalar(),
				unit:        s.Amount.UnitType.Symbol,
			})
		}
	}
	return i
}
