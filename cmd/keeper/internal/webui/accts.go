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
	"sort"

	"go.felesatra.moe/keeper/journal"
)

func journalAccounts(j *journal.Journal) []journal.Account {
	var accounts []journal.Account
	for a := range j.AccountEntries {
		accounts = append(accounts, a)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i] < accounts[j] })
	return accounts
}

func revenueAccounts(a []journal.Account) []journal.Account {
	var a2 []journal.Account
	for _, a := range a {
		if a.Under("Revenues") {
			a2 = append(a2, a)
		}
	}
	return a2
}

func expenseAccounts(a []journal.Account) []journal.Account {
	var a2 []journal.Account
	for _, a := range a {
		if a.Under("Expenses") {
			a2 = append(a2, a)
		}
	}
	return a2
}
