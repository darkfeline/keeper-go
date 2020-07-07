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
	"html/template"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/internal/month"
	"go.felesatra.moe/keeper/journal"
)

func NewHandler(o []journal.Option) http.Handler {
	h := handler{o}
	m := http.NewServeMux()
	m.HandleFunc("/", h.handleIndex)
	m.HandleFunc("/style.css", h.handleStyle)
	m.HandleFunc("/CollapsibleLists.js", h.handleCollapse)
	m.HandleFunc("/accounts", h.handleAccounts)
	m.HandleFunc("/trial", h.handleTrial)
	m.HandleFunc("/income", h.handleIncome)
	m.HandleFunc("/capital", h.handleCapital)
	m.HandleFunc("/balance", h.handleBalance)
	m.HandleFunc("/cash", h.handleCash)
	m.HandleFunc("/ledger", h.handleLedger)
	return m
}

type handler struct {
	o []journal.Option
}

func (h handler) handleIndex(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	d := indexData{
		BalanceErrors: j.BalanceErrors,
	}
	execute(w, indexTemplate, d)
}

func (h handler) handleAccounts(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	d := accountsData{
		AccountTree: journalAccountTree(j),
	}
	execute(w, accountsTemplate, d)
}

func (h handler) handleStyle(w http.ResponseWriter, req *http.Request) {
	http.ServeContent(w, req, "style.css", time.Time{}, strings.NewReader(styleText))
}

func (h handler) handleCollapse(w http.ResponseWriter, req *http.Request) {
	http.ServeContent(w, req, "CollapsibleLists.js", time.Time{}, strings.NewReader(collapseText))
}

func (h handler) handleTrial(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	r, t := makeTrialRows(j.Accounts(), j.Balances)
	r = append(r, t.Rows("Total")...)
	d := trialData{Rows: r}
	execute(w, trialTemplate, d)
}

func (h handler) handleIncome(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compile(journal.Ending(end))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	a := j.Accounts()
	b := j.Balances
	d := stmtData{
		Title: "Income Statement",
		Month: month.Format(end),
	}
	add := func(r ...stmtRow) {
		d.Rows = append(d.Rows, r...)
	}

	add(stmtRow{Description: "Income", Section: true})
	// Income is credit balance.
	b.Neg()
	r, rt := makeStmtRows(revenueAccounts(a), b)
	add(r...)
	add(makeStmtBalance("Total Income", rt)...)

	add(stmtRow{Description: "Expenses", Section: true})
	// Expenses are debit balance.
	b.Neg()
	r, et := makeStmtRows(expenseAccounts(a), b)
	add(r...)
	add(makeStmtBalance("Total Expenses", et)...)

	add(stmtRow{Description: "Net Profit", Section: true})
	add(makeStmtBalance("Total Net Profit", rt)...)
	execute(w, stmtTemplate, d)
}

func (h handler) handleCapital(w http.ResponseWriter, req *http.Request) {
	panic("Not implemented")
}

func (h handler) handleBalance(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compile(journal.Ending(end))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	a := j.Accounts()
	b := j.Balances
	d := stmtData{
		Title: "Balance Sheet",
		Month: month.Format(end),
	}
	add := func(r ...stmtRow) {
		d.Rows = append(d.Rows, r...)
	}

	add(stmtRow{Description: "Assets", Section: true})
	// Assets are debit balance.
	r, t := makeStmtRows(assetAccounts(a), b)
	add(r...)
	add(makeStmtBalance("Total Assets", t)...)

	add(stmtRow{Description: "Liabilities", Section: true})
	// Liabilities are credit balance.
	b.Neg()
	r, lt := makeStmtRows(liabilityAccounts(a), b)
	add(r...)
	add(makeStmtBalance("Total Liabilities", lt)...)

	add(stmtRow{Description: "Equity", Section: true})
	// Equity is credit balance.
	r, et := makeStmtRows(equityAccounts(a), b)
	add(r...)
	add(makeStmtBalance("Total Equity", et)...)

	for _, a := range et.Amounts() {
		lt.Add(a)
	}
	add(stmtRow{})
	add(makeStmtBalance("Total Liabilities & Equity", lt)...)

	execute(w, stmtTemplate, d)
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
		http.Error(w, err.Error(), 500)
		return
	}
	d := ledgerData{Account: account}
	b := make(journal.Balance)
	for _, e := range accountEntries(j.Entries, account) {
		d.Rows = append(d.Rows, makeLedgerRows(b, e, account)...)
	}
	execute(w, ledgerTemplate, d)
}

func (h handler) compile(o ...journal.Option) (*journal.Journal, error) {
	var o2 []journal.Option
	o2 = append(o2, h.o...)
	o2 = append(o2, o...)
	return journal.Compile(o2...)
}

func getQueryAccount(req *http.Request) journal.Account {
	v := req.URL.Query()["account"]
	if len(v) == 0 {
		return ""
	}
	return journal.Account(v[0])
}

func execute(w http.ResponseWriter, t *template.Template, data interface{}) {
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(b.Bytes())
}
