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
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/internal/account"
	"go.felesatra.moe/keeper/internal/month"
	"go.felesatra.moe/keeper/internal/webui/templates"
	"go.felesatra.moe/keeper/journal"
)

func NewHandler(c *account.Classifier, o []journal.Option) http.Handler {
	h := handler{
		c: c,
		o: o,
	}
	m := http.NewServeMux()
	m.HandleFunc("/", h.handleIndex)
	m.HandleFunc("/style.css", h.handleStyle)
	m.HandleFunc("/trial", h.handleTrial)
	m.HandleFunc("/income", h.handleIncome)
	m.HandleFunc("/capital", h.handleCapital)
	m.HandleFunc("/balance", h.handleBalance)
	m.HandleFunc("/cash", h.handleCash)
	m.HandleFunc("/ledger", h.handleLedger)
	return m
}

type handler struct {
	c *account.Classifier
	o []journal.Option
}

func (h handler) handleIndex(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		h.writeError(w, err)
		return
	}
	d := templates.IndexData{
		BalanceErrors: j.BalanceErrors,
	}
	h.execute(w, templates.Index, d)
}

func (h handler) handleStyle(w http.ResponseWriter, req *http.Request) {
	http.ServeContent(w, req, "style.css", time.Time{}, strings.NewReader(templates.StyleText))
}

func (h handler) handleTrial(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		h.writeError(w, err)
		return
	}
	r, t := makeTrialRows(j.Accounts(), j.Balances)
	r = append(r, t.Rows("Total")...)
	d := templates.TrialData{Rows: r}
	h.execute(w, templates.Trial, d)
}

func (h handler) handleIncome(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compile(journal.Ending(end))
	if err != nil {
		h.writeError(w, err)
		return
	}
	c, err := h.classifier()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	a := j.Accounts()
	b := j.Balances
	d := templates.StmtData{
		Title: "Income Statement",
		Month: month.Format(end),
	}
	add := func(r ...templates.StmtRow) {
		d.Rows = append(d.Rows, r...)
	}

	add(templates.StmtRow{Description: "Income", Section: true})
	// Income is credit balance.
	b.Neg()
	r, rt := makeStmtRows(filter(a, c.IsIncome), b)
	add(r...)
	add(makeStmtBalance("Total Income", rt)...)

	add(templates.StmtRow{Description: "Expenses", Section: true})
	// Expenses are debit balance.
	b.Neg()
	r, et := makeStmtRows(filter(a, c.IsExpenses), b)
	add(r...)
	add(makeStmtBalance("Total Expenses", et)...)

	for _, am := range et.Amounts() {
		rt.Add(am.Neg())
	}
	add(templates.StmtRow{Description: "Net Profit", Section: true})
	add(makeStmtBalance("Total Net Profit", rt)...)
	h.execute(w, templates.Stmt, d)
}

func (h handler) handleCapital(w http.ResponseWriter, req *http.Request) {
	panic("Not implemented")
}

func (h handler) handleBalance(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compile(journal.Ending(end))
	if err != nil {
		h.writeError(w, err)
		return
	}
	c, err := h.classifier()
	if err != nil {
		h.writeError(w, err)
		return
	}

	a := j.Accounts()
	b := j.Balances
	d := templates.StmtData{
		Title: "Balance Sheet",
		Month: month.Format(end),
	}
	add := func(r ...templates.StmtRow) {
		d.Rows = append(d.Rows, r...)
	}

	add(templates.StmtRow{Description: "Assets", Section: true})
	// Assets are debit balance.
	r, t := makeStmtRows(filter(a, c.IsAssets), b)
	add(r...)
	add(makeStmtBalance("Total Assets", t)...)

	add(templates.StmtRow{Description: "Liabilities", Section: true})
	// Liabilities are credit balance.
	b.Neg()
	r, lt := makeStmtRows(filter(a, c.IsLiabilities), b)
	add(r...)
	add(makeStmtBalance("Total Liabilities", lt)...)

	add(templates.StmtRow{Description: "Equity", Section: true})
	// Equity is credit balance.
	r, et := makeStmtRows(filter(a, func(a journal.Account) bool {
		return c.IsEquity(a) ||
			c.IsIncome(a) ||
			c.IsExpenses(a) ||
			c.IsTrading(a)
	}), b)
	add(r...)
	add(makeStmtBalance("Total Equity", et)...)

	for _, a := range et.Amounts() {
		lt.Add(a)
	}
	add(templates.StmtRow{})
	add(makeStmtBalance("Total Liabilities & Equity", lt)...)

	h.execute(w, templates.Stmt, d)
}

func (h handler) handleCash(w http.ResponseWriter, req *http.Request) {
	panic("Not implemented")
}

func getQueryMonth(req *http.Request) civil.Date {
	v := req.URL.Query()["month"]
	if len(v) == 0 {
		return civil.DateOf(time.Now())
	}
	d, err := month.Parse(v[0])
	if err != nil {
		return civil.DateOf(time.Now())
	}
	return d
}

func (h handler) handleLedger(w http.ResponseWriter, req *http.Request) {
	account := getQueryAccount(req)
	j, err := h.compile()
	if err != nil {
		h.writeError(w, err)
		return
	}
	d := templates.LedgerData{Account: account}
	b := make(journal.Balance)
	for _, e := range accountEntries(j.Entries, account) {
		d.Rows = append(d.Rows, makeLedgerRows(b, e, account)...)
	}
	h.execute(w, templates.Ledger, d)
}

func (h handler) compile(o ...journal.Option) (*journal.Journal, error) {
	var o2 []journal.Option
	o2 = append(o2, h.o...)
	o2 = append(o2, o...)
	return journal.Compile(o2...)
}

func (h handler) classifier() (*account.Classifier, error) {
	return h.c, nil
}

func (h handler) writeError(w http.ResponseWriter, err error) {
	msg := fmt.Sprintf("Error: %s\n\nDebug info:\n\nhandler: %#v", err, h)
	http.Error(w, msg, 500)
}

func (h handler) execute(w http.ResponseWriter, t *template.Template, data interface{}) {
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		h.writeError(w, err)
		return
	}
	w.Write(b.Bytes())
}

func getQueryAccount(req *http.Request) journal.Account {
	v := req.URL.Query()["account"]
	if len(v) == 0 {
		return ""
	}
	return journal.Account(v[0])
}

type totalBalance struct {
	Debit  journal.Balance
	Credit journal.Balance
}

func (t totalBalance) Rows(desc string) []templates.TrialRow {
	var r []templates.TrialRow
	for i, u := range balanceUnits(t.Debit, t.Credit) {
		e := templates.TrialRow{
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

func makeTrialRows(a []journal.Account, b journal.Balances) ([]templates.TrialRow, totalBalance) {
	t := totalBalance{
		Debit:  make(journal.Balance),
		Credit: make(journal.Balance),
	}
	var r []templates.TrialRow
	for _, a := range a {
		for i, amt := range b[a].Amounts() {
			e := templates.TrialRow{}
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

func makeStmtRows(a []journal.Account, b journal.Balances) ([]templates.StmtRow, journal.Balance) {
	t := make(journal.Balance)
	var r []templates.StmtRow
	for _, a := range a {
		for i, amt := range b[a].Amounts() {
			e := templates.StmtRow{Amount: amt}
			if i == 0 {
				e.Description = string(a)
				e.Account = true
			}
			r = append(r, e)
			t.Add(amt)
		}
	}
	return r, t
}

func makeStmtBalance(desc string, b journal.Balance) []templates.StmtRow {
	var r []templates.StmtRow
	for i, u := range balanceUnits(b) {
		e := templates.StmtRow{Amount: b.Amount(u)}
		if i == 0 {
			e.Description = desc
		}
		r = append(r, e)
	}
	return r
}

func makeLedgerRows(b journal.Balance, e journal.Entry, a journal.Account) []templates.LedgerRow {
	switch e := e.(type) {
	case *journal.Transaction:
		return convertTransaction(b, a, e)
	case *journal.BalanceAssert:
		return convertBalance(e)
	case *journal.DisableAccount:
		return []templates.LedgerRow{{
			Entry:       e,
			Description: "(disabled)",
		}}
	default:
		panic(fmt.Sprintf("unknown entry %T", e))
	}
}

func convertBalance(e *journal.BalanceAssert) []templates.LedgerRow {
	units := balanceUnits(e.Actual, e.Declared, e.Diff)
	var entries []templates.LedgerRow
	for _, u := range units {
		le := templates.LedgerRow{
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

func convertTransaction(b journal.Balance, a journal.Account, e *journal.Transaction) []templates.LedgerRow {
	var entries []templates.LedgerRow
	first := true
	for _, s := range e.Splits {
		if s.Account != a {
			continue
		}
		le := templates.LedgerRow{
			Amount: s.Amount,
		}
		b.Add(s.Amount)
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
	amts := b.Amounts()
	if len(amts) == 0 {
		return entries
	}
	entries[len(entries)-1].Balance = amts[0]
	for _, a := range amts[1:] {
		le := templates.LedgerRow{
			Balance: a,
		}
		entries = append(entries, le)
	}
	return entries
}
