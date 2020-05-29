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
	"flag"
	"fmt"
	"os"

	"go.felesatra.moe/keeper/journal"
)

var checkCmd = &command{
	name: "check",
	run: func(cmd *command, args []string) {
		fs := flag.NewFlagSet(cmd.name, flag.ExitOnError)
		if err := fs.Parse(args); err != nil {
			panic(err)
		}
		var o []journal.Option
		for _, f := range fs.Args() {
			o = append(o, journal.File(f))
		}
		j, err := journal.Compile(o...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "keeper: %s\n", err)
			os.Exit(1)
		}
		if len(j.BalanceErrors) > 0 {
			for _, e := range j.BalanceErrors {
				fmt.Fprintf(os.Stderr, "%s %s %s declared %s, actual %s (diff %s)\n",
					e.EntryPos, e.EntryDate, e.Account,
					e.Declared, e.Actual, e.Diff)
			}
			fmt.Fprintf(os.Stderr, "keeper: balance errors\n")
			os.Exit(1)
		}
	},
}
