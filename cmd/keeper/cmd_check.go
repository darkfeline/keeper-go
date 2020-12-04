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
	"os"

	"go.felesatra.moe/keeper/journal"
)

var checkCmd = &command{
	usageLine: "check [files]",
	run: func(cmd *command, args []string) {
		fs := cmd.flagSet()
		fs.Parse(args)
		o := []journal.Option{journal.File(fs.Args()...)}
		j, err := journal.Compile(o...)
		if err != nil {
			errf("%s", err)
			os.Exit(1)
		}
		checkBalanceErrsAndExit(j)
	},
}

func checkBalanceErrsAndExit(j *journal.Journal) {
	if len(j.BalanceErrors) > 0 {
		for _, e := range j.BalanceErrors {
			errf("%s %s %s declared %s, actual %s (diff %s)",
				e.EntryPos, e.EntryDate, e.Account,
				e.Declared, e.Actual, e.Diff)
		}
		errf("balance errors")
		os.Exit(1)
	}
}
