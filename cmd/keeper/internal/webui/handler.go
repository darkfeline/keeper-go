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

	"go.felesatra.moe/keeper/journal"
)

func NewHandler(o []journal.Option) http.Handler {
	h := handler{o}
	m := http.NewServeMux()
	m.HandleFunc("/", h.handleIndex)
	m.HandleFunc("/style.css", h.handleStyle)
	m.HandleFunc("/accounts", h.handleAccounts)
	m.HandleFunc("/trial", h.handleTrial)
	m.HandleFunc("/income", h.handleIncome)
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

func (h handler) handleTrial(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	r, t := makeBalanceRows(journalAccounts(j), j.Balances)
	r = append(r, t.Rows("Total")...)
	d := trialData{Entries: r}
	execute(w, trialTemplate, d)
}

func (h handler) handleIncome(w http.ResponseWriter, req *http.Request) {
	panic("Not implemented")
}

func (h handler) handleLedger(w http.ResponseWriter, req *http.Request) {
	account := getQueryAccount(req)
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	d := ledgerData{Account: account}
	for _, e := range j.AccountEntries[account] {
		d.Entries = append(d.Entries, makeLedgerRows(e, account)...)
	}
	execute(w, ledgerTemplate, d)
}

func (h handler) compile() (*journal.Journal, error) {
	return journal.Compile(h.o...)
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
