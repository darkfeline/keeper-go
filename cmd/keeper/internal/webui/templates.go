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
	"html/template"

	"go.felesatra.moe/keeper/journal"
)

//go:generate binpack -name baseText base.html

var baseTemplate = template.Must(template.New("base").Parse(baseText))

type baseData struct {
	Title string
	Body  template.HTML
}

//go:generate binpack -name indexText index.html

var indexTemplate = template.Must(clone(baseTemplate).Parse(indexText))

type indexData struct {
	BalanceErrors []journal.BalanceAssert
}

func (indexData) Title() string { return "" }

//go:generate binpack -name accountsText accounts.html

var accountsTemplate = template.Must(clone(baseTemplate).Parse(accountsText))

type accountsData struct {
	AccountTree *accountTree
}

func (accountsData) Title() string { return "Accounts" }

//go:generate binpack -name trialText trial.html

var trialTemplate = template.Must(clone(baseTemplate).Parse(trialText))

type trialData struct {
	Account journal.Account
	Entries []ledgerEntry
}

func (trialData) Title() string { return "Trial Balance" }

//go:generate binpack -name ledgerText ledger.html

var ledgerTemplate = template.Must(clone(baseTemplate).Parse(ledgerText))

type ledgerData struct {
	Account journal.Account
	Entries []ledgerEntry
}

func (d ledgerData) Title() string {
	return fmt.Sprintf("Ledger for %s", d.Account)
}

type ledgerEntry struct {
	journal.Entry
	Description string
	Amount      journal.Amount
	Balance     journal.Balance
}

func convertEntry(e journal.Entry, a journal.Account) []ledgerEntry {
	switch e := e.(type) {
	case journal.Transaction:
		return convertTransaction(e, a)
	case journal.BalanceAssert:
		le := ledgerEntry{
			Entry:   e,
			Balance: e.Actual,
		}
		if e.Diff.Empty() {
			le.Description = "(balance)"
		} else {
			le.Description = fmt.Sprintf("(balance error, declared %s, diff %s)",
				e.Declared, e.Actual)
		}
		return []ledgerEntry{le}
	default:
		panic(fmt.Sprintf("unknown entry %t", e))
	}
}

func convertTransaction(e journal.Transaction, a journal.Account) []ledgerEntry {
	var entries []ledgerEntry
	for _, s := range e.Splits {
		if s.Account != a {
			continue
		}
		le := ledgerEntry{
			Entry:       e,
			Description: e.Description,
			Amount:      s.Amount,
		}
		entries = append(entries, le)
	}
	if len(entries) == 0 {
		return entries
	}
	entries[len(entries)-1].Balance = e.Balances[a]
	return entries
}

func clone(t *template.Template) *template.Template {
	return template.Must(t.Clone())
}
