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

package webui

import (
	"fmt"

	"go.felesatra.moe/keeper/chart"
	"go.felesatra.moe/keeper/journal"
)

func accountEntries(e []journal.Entry, a journal.Account) []journal.Entry {
	return filterEntries(e, func(a2 journal.Account) bool {
		return a == a2
	})
}

func filterEntries(e []journal.Entry, f func(journal.Account) bool) []journal.Entry {
	var e2 []journal.Entry
	for _, e := range e {
		switch e := e.(type) {
		case *journal.Transaction:
			for _, s := range e.Splits {
				if f(s.Account) {
					e2 = append(e2, e)
					break
				}
			}
		case *journal.BalanceAssert:
			if f(e.Account) {
				e2 = append(e2, e)
			}
		case *journal.DisableAccount:
			if f(e.Account) {
				e2 = append(e2, e)
			}
		default:
			panic(fmt.Sprintf("unknown entry %T", e))
		}
	}
	return e2
}

func revenueAccounts(a []journal.Account) []journal.Account {
	c := chart.New(a)
	return c.Income()
}

func expenseAccounts(a []journal.Account) []journal.Account {
	c := chart.New(a)
	return c.Expenses()
}

func assetAccounts(a []journal.Account) []journal.Account {
	c := chart.New(a)
	return c.Assets()
}

func liabilityAccounts(a []journal.Account) []journal.Account {
	c := chart.New(a)
	return c.Liabilities()
}

func equityAccounts(a []journal.Account) []journal.Account {
	c := chart.New(a)
	return c.Equity()
}
