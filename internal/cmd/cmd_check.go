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

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check [file]",
	Short: "check file for errors",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		j, err := compileFile(args[0])
		if err != nil {
			return err
		}
		if len(j.BalanceErrors) > 0 {
			for _, e := range j.BalanceErrors {
				fmt.Fprintf(os.Stderr, "%s %s %s declared %s, actual %s (diff %s)\n",
					e.EntryPos, e.EntryDate, e.Account,
					e.Declared, e.Actual, e.Diff)
			}
			return errors.New("journal has balance errors")
		}
		return nil
	},
}
