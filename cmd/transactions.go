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

	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse"
	"go.felesatra.moe/keeper/report"
)

func init() {
	rootCmd.AddCommand(splitsCmd)
}

var splitsCmd = &cobra.Command{
	Use:   "splits",
	Short: "Print splits for account",
	Args:  cobra.ExactArgs(2),
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
		ts = report.AccountSplits(ts, book.Account(args[1]))
		writeSplits(os.Stdout, ts)
		return nil
	},
}

func writeSplits(w io.Writer, ts []book.Transaction) error {
	bw := bufio.NewWriter(w)
	// XXXXXXXXXXXXXXXXXX
	return bw.Flush()
}
