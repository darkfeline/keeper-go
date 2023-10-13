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

// Package templates contains templates for the Web UI.
package templates

import (
	"embed"
	"fmt"
	"html/template"

	_ "embed"

	"go.felesatra.moe/keeper/journal"
	"go.felesatra.moe/keeper/reports"
)

//go:embed style.css
var StyleText []byte

//go:embed *.html
var f embed.FS

var Base = template.Must(template.ParseFS(f, "base.html"))

type BaseData struct {
	Title string
	Body  template.HTML
}

func extendBase(file string) *template.Template {
	return template.Must(clone(Base).ParseFS(f, file))
}

var Index = extendBase("index.html")

type IndexData struct {
	BalanceErrors []*journal.BalanceAssert
}

func (IndexData) Title() string { return "" }

var Accounts = extendBase("accounts.html")

type AccountsData struct {
	Accounts []struct {
		Account journal.Account
		Empty   bool
	}
	Disabled []journal.Account
}

func (AccountsData) Title() string { return "" }

var Trial = extendBase("trial.html")

type TrialData struct {
	Rows []TrialRow
}

func (TrialData) Title() string { return "Trial Balance" }

type TrialRow struct {
	Account   string
	DebitBal  *journal.Amount
	CreditBal *journal.Amount
}

var Stmt = extendBase("stmt.html")

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
	Amount  *journal.Amount
}

var Ledger = extendBase("ledger.html")

type LedgerData struct {
	Account journal.Account
	Rows    []LedgerRow
}

func (d LedgerData) Title() string {
	return fmt.Sprintf("Ledger for %s", d.Account)
}

type LedgerRow struct {
	Date        string
	Description string
	Ref         string
	Pair        reports.Pair[*journal.Amount]
	Balance     *journal.Amount
}

func clone(t *template.Template) *template.Template {
	return template.Must(t.Clone())
}
