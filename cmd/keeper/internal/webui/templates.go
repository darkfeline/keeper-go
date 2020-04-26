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
	"sort"

	"go.felesatra.moe/keeper/journal"
)

//go:generate binpack -name baseText base.html
//go:generate binpack -name styleText style.css

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
	Rows []trialRow
}

func (trialData) Title() string { return "Trial Balance" }

type trialRow struct {
	Account   string
	DebitBal  journal.Amount
	CreditBal journal.Amount
}

type totalBalance struct {
	Debit  journal.Balance
	Credit journal.Balance
}

func (t totalBalance) Rows(desc string) []trialRow {
	var r []trialRow
	for i, u := range balanceUnits(t.Debit, t.Credit) {
		e := trialRow{
			DebitBal:  t.Debit.Amount(u),
			CreditBal: t.Credit.Amount(u),
		}
		if i == 0 {
			e.Account = desc
		}
		r = append(r, e)
	}
	return r
}

func makeTrialRows(a []journal.Account, b journal.Balances) ([]trialRow, totalBalance) {
	t := totalBalance{
		Debit:  make(journal.Balance),
		Credit: make(journal.Balance),
	}
	var r []trialRow
	for _, a := range a {
		for i, amt := range b[a].Amounts() {
			e := trialRow{}
			if amt.Number > 0 {
				e.DebitBal = amt
				t.Debit.Add(amt)
			} else {
				e.CreditBal = amt
				t.Credit.Add(amt)
			}
			if i == 0 {
				e.Account = string(a)
			}
			r = append(r, e)
		}
	}
	return r, t
}

//go:generate binpack -name stmtText stmt.html

var stmtTemplate = template.Must(clone(baseTemplate).Parse(stmtText))

type stmtData struct {
	Title string
	Rows  []stmtRow
}

type stmtRow struct {
	Description string
	Section     bool
	Amount      journal.Amount
}

func makeStmtRows(a []journal.Account, b journal.Balances) ([]stmtRow, journal.Balance) {
	t := make(journal.Balance)
	var r []stmtRow
	for _, a := range a {
		for i, amt := range b[a].Amounts() {
			e := stmtRow{Amount: amt}
			if i == 0 {
				e.Description = string(a)
			}
			r = append(r, e)
			t.Add(amt)
		}
	}
	return r, t
}

func makeStmtBalance(desc string, b journal.Balance) []stmtRow {
	var r []stmtRow
	for i, u := range balanceUnits(b) {
		e := stmtRow{Amount: b.Amount(u)}
		if i == 0 {
			e.Description = desc
		}
		r = append(r, e)
	}
	return r
}

//go:generate binpack -name ledgerText ledger.html

var ledgerTemplate = template.Must(clone(baseTemplate).Parse(ledgerText))

type ledgerData struct {
	Account journal.Account
	Rows    []ledgerRow
}

func (d ledgerData) Title() string {
	return fmt.Sprintf("Ledger for %s", d.Account)
}

type ledgerRow struct {
	Entry       journal.Entry
	Description string
	Amount      journal.Amount
	Balance     journal.Amount
}

func (e ledgerRow) Position() string {
	if e.Entry == nil {
		return ""
	}
	return e.Entry.Position().String()
}

func (e ledgerRow) Date() string {
	if e.Entry == nil {
		return ""
	}
	return e.Entry.Date().String()
}

func makeLedgerRows(e journal.Entry, a journal.Account) []ledgerRow {
	switch e := e.(type) {
	case journal.Transaction:
		return convertTransaction(e, a)
	case journal.BalanceAssert:
		return convertBalance(e)
	default:
		panic(fmt.Sprintf("unknown entry %t", e))
	}
}

func convertBalance(e journal.BalanceAssert) []ledgerRow {
	units := balanceUnits(e.Actual, e.Declared, e.Diff)
	var entries []ledgerRow
	for _, u := range units {
		le := ledgerRow{
			Entry:   e,
			Balance: e.Actual.Amount(u),
		}
		if e.Diff[u] == 0 {
			le.Description = "(balance)"
		} else {
			le.Description = fmt.Sprintf("(balance error, declared %s, diff %s)",
				e.Declared.Amount(u), e.Diff.Amount(u))
		}
		entries = append(entries, le)
	}
	return entries
}

// balanceUnits returns all of the units in the balances.
func balanceUnits(b ...journal.Balance) []journal.Unit {
	seen := make(map[journal.Unit]bool)
	for _, b := range b {
		for u := range b {
			seen[u] = true
		}
	}
	var units []journal.Unit
	for u, v := range seen {
		if v {
			units = append(units, u)
		}
	}
	sort.Slice(units, func(i, j int) bool { return units[i].Symbol < units[j].Symbol })
	return units
}

func convertTransaction(e journal.Transaction, a journal.Account) []ledgerRow {
	var entries []ledgerRow
	first := true
	for _, s := range e.Splits {
		if s.Account != a {
			continue
		}
		le := ledgerRow{
			Amount: s.Amount,
		}
		if first {
			le.Entry = e
			le.Description = e.Description
			first = false
		}
		entries = append(entries, le)
	}
	if len(entries) == 0 {
		return entries
	}
	amts := e.Balances[a].Amounts()
	if len(amts) == 0 {
		return entries
	}
	entries[len(entries)-1].Balance = amts[0]
	for _, a := range amts[1:] {
		le := ledgerRow{
			Balance: a,
		}
		entries = append(entries, le)
	}
	return entries
}

func clone(t *template.Template) *template.Template {
	return template.Must(t.Clone())
}
