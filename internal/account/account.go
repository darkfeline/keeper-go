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

// Package account contains account related utilities.
package account

import (
	"fmt"

	"go.felesatra.moe/keeper/journal"
)

func FilterEntries(e []journal.Entry, a journal.Account) []journal.Entry {
	var e2 []journal.Entry
	for _, e := range e {
		switch e := e.(type) {
		case *journal.Transaction:
			for _, s := range e.Splits {
				if s.Account == a {
					e2 = append(e2, e)
					break
				}
			}
		case *journal.BalanceAssert:
			if e.Account == a {
				e2 = append(e2, e)
			}
		case *journal.DisableAccount:
			if e.Account == a {
				e2 = append(e2, e)
			}
		default:
			panic(fmt.Sprintf("unknown entry %T", e))
		}
	}
	return e2
}
