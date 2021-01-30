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

// Package templates contains templates for the web UI.
package templates

import (
	"fmt"
	"html/template"

	"go.felesatra.moe/keeper/journal"
)

//go:generate binpack -name StyleText style.css

//go:generate binpack -name baseText base.html

var Base = template.Must(template.New("base").Parse(baseText))

type BaseData struct {
	Title string
	Body  template.HTML
}

//go:generate binpack -name indexText index.html

var Index = template.Must(clone(Base).Parse(indexText))

type IndexData struct {
	BalanceErrors []*journal.BalanceAssert
}

func (IndexData) Title() string { return "" }

//go:generate binpack -name accountsText accounts.html

var Accounts = template.Must(clone(Base).Parse(accountsText))

type AccountsData struct {
	Accounts []journal.Account
	Disabled []journal.Account
}

func (AccountsData) Title() string { return "" }

//go:generate binpack -name trialText trial.html

var Trial = template.Must(clone(Base).Parse(trialText))

type TrialData struct {
	Rows []TrialRow
}

func (TrialData) Title() string { return "Trial Balance" }

type TrialRow struct {
	Account   string
	DebitBal  journal.Amount
	CreditBal journal.Amount
}

//go:generate binpack -name stmtText stmt.html

var Stmt = template.Must(clone(Base).Parse(stmtText))

type StmtData struct {
	Title string
	Month string // YYYY-MM
	Rows  []StmtRow
}

type StmtRow struct {
	Description string
	// Indicates the row is a section header, giving it emphasis.
	Section bool
	// Indicates the description is an account name and makes it a
	// link to the account's ledger page.
	Account bool
	Amount  journal.Amount
}

//go:generate binpack -name ledgerText ledger.html

var Ledger = template.Must(clone(Base).Parse(ledgerText))

type LedgerData struct {
	Account journal.Account
	Rows    []LedgerRow
}

func (d LedgerData) Title() string {
	return fmt.Sprintf("Ledger for %s", d.Account)
}

type LedgerRow struct {
	Entry       journal.Entry
	Description string
	Amount      journal.Amount
	Balance     journal.Amount
}

func (e LedgerRow) Position() string {
	if e.Entry == nil {
		return ""
	}
	return e.Entry.Position().String()
}

func (e LedgerRow) Date() string {
	if e.Entry == nil {
		return ""
	}
	return e.Entry.Date().String()
}

func clone(t *template.Template) *template.Template {
	return template.Must(t.Clone())
}
