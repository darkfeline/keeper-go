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

// Package webui implements the Web UI for keeper.
package webui

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/internal/config"
	"go.felesatra.moe/keeper/internal/month"
	"go.felesatra.moe/keeper/internal/webui/templates"
	"go.felesatra.moe/keeper/journal"
	"go.felesatra.moe/keeper/reports"
)

func NewHandler(configPath string, a *journal.CompileArgs) http.Handler {
	h := handler{
		configPath: configPath,
		a:          a,
	}
	m := http.NewServeMux()
	m.HandleFunc("/", h.handleIndex)
	m.HandleFunc("/style.css", h.handleStyle)
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
	configPath string
	a          *journal.CompileArgs
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
	http.ServeContent(w, req, "style.css", time.Time{}, bytes.NewReader(templates.StyleText))
}

func (h handler) handleAccounts(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		h.writeError(w, err)
		return
	}
	d := makeAccountsData(j)
	h.execute(w, templates.Accounts, d)
}

func (h handler) handleTrial(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		h.writeError(w, err)
		return
	}
	d := makeTrialData(reports.NewTrialBalance(j))
	h.execute(w, templates.Trial, d)
}

func (h handler) handleIncome(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compileEnding(end)
	if err != nil {
		h.writeError(w, err)
		return
	}
	c, err := h.config()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	a := sortedAccounts(j)
	b := j.Balances
	s := stmt{
		StmtData: &templates.StmtData{
			Title: "Income Statement",
			Month: month.Format(end),
		},
	}

	s.addSection("Income")
	// Income is credit balance.
	b.Neg()
	for _, a := range a {
		if c.IsIncome(a) {
			s.addAccount(a, b[a])
		}
	}
	var income journal.Balance
	income.Set(&s.bal)
	s.addTotal("Total Income")

	s.addSection("Expenses")
	// Expenses are debit balance.
	b.Neg()
	for _, a := range a {
		if c.IsExpenses(a) {
			s.addAccount(a, b[a])
		}
	}
	var expenses journal.Balance
	expenses.Set(&s.bal)
	s.addTotal("Total Expenses")

	expenses.Neg()
	income.AddBal(&expenses)
	s.addSection("Net Profit")
	s.addBalanceRows(templates.StmtRow{Description: "Total Net Profit"}, &income)
	h.execute(w, templates.Stmt, s.StmtData)
}

func (h handler) handleCapital(w http.ResponseWriter, req *http.Request) {
	h.writeError(w, errors.New("not implemented"))
}

func (h handler) handleBalance(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compileEnding(end)
	if err != nil {
		h.writeError(w, err)
		return
	}
	c, err := h.config()
	if err != nil {
		h.writeError(w, err)
		return
	}

	a := sortedAccounts(j)
	b := j.Balances
	s := stmt{
		StmtData: &templates.StmtData{
			Title: "Balance Sheet",
			Month: month.Format(end),
		},
	}

	s.addSection("Assets")
	// Assets are debit balance.
	for _, a := range a {
		if c.IsAssets(a) {
			s.addAccount(a, b[a])
		}
	}
	s.addTotal("Total Assets")

	s.addSection("Liabilities")
	// Liabilities are credit balance.
	b.Neg()
	for _, a := range a {
		if c.IsLiabilities(a) {
			s.addAccount(a, b[a])
		}
	}
	var liabilities journal.Balance
	liabilities.Set(&s.bal)
	s.addTotal("Total Liabilities")

	s.addSection("Equity")
	// Equity is credit balance.
	for _, a := range a {
		if c.IsEquity(a) || c.IsIncome(a) || c.IsExpenses(a) || c.IsTrading(a) {
			s.addAccount(a, b[a])
		}
	}
	var equity journal.Balance
	equity.Set(&s.bal)
	s.addTotal("Total Equity")

	equity.AddBal(&liabilities)
	s.addRows(templates.StmtRow{})
	s.addBalanceRows(templates.StmtRow{Description: "Total Liabilities & Equity"}, &equity)

	h.execute(w, templates.Stmt, s.StmtData)
}

func (h handler) handleCash(w http.ResponseWriter, req *http.Request) {
	end := month.LastDay(getQueryMonth(req))
	j, err := h.compileEnding(end)
	if err != nil {
		h.writeError(w, err)
		return
	}
	c, err := h.config()
	if err != nil {
		h.writeError(w, err)
		return
	}

	start := month.FirstDay(end)
	a := cashAccounts(c, sortedAccounts(j))
	delta := accountDeltas(j.Entries, a, start)
	// We want to represent flow away from cash accounts.
	delta.Neg()
	type fl struct {
		a  journal.Account
		am *journal.Amount
	}
	var infl, outfl []fl
	for _, a := range delta.Accounts() {
		if c.IsCash(a) {
			continue
		}
		for _, am := range delta[a].Amounts() {
			switch am.Number.Sign() {
			case -1:
				outfl = append(outfl, fl{a, am})
			case 1:
				infl = append(infl, fl{a, am})
			}
		}
	}

	s := stmt{
		StmtData: &templates.StmtData{
			Title: "Cash Flow",
			Month: month.Format(end),
		},
	}
	s.addSection("Starting Balances")
	starting := j.BalancesEnding(start)
	for _, a := range a {
		s.addAccount(a, starting[a])
	}
	s.addTotal("Total Starting")

	s.addSection("Inflow")
	for _, v := range infl {
		s.addAccountAmount(v.a, v.am)
	}
	s.addTotal("Total Inflow")

	s.addSection("Outflow")
	for _, v := range outfl {
		s.addAccountAmount(v.a, v.am)
	}
	s.addTotal("Total Outflow")

	s.addSection("Ending Balances")
	for _, a := range a {
		s.addAccount(a, j.Balances[a])
	}
	s.addTotal("Total Ending")

	h.execute(w, templates.Stmt, s.StmtData)
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
	a := getQueryAccount(req)
	j, err := h.compile()
	if err != nil {
		h.writeError(w, err)
		return
	}
	l := reports.NewAccountLedger(j, a)
	d := makeLedgerData(l)
	h.execute(w, templates.Ledger, d)
}

func (h handler) compile() (*journal.Journal, error) {
	return journal.Compile(h.a)
}

func (h handler) compileEnding(d civil.Date) (*journal.Journal, error) {
	a2 := *h.a
	a2.Ending = d
	return journal.Compile(&a2)
}

func (h handler) config() (*config.Config, error) {
	c := &config.Config{}
	if h.configPath == "" {
		return c, nil
	}
	f, err := os.Open(h.configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = config.Load(c, f)
	return c, err
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

func makeAccountsData(j *journal.Journal) templates.AccountsData {
	var d templates.AccountsData
	for _, a := range sortedAccounts(j) {
		if j.Accounts[a].Disabled == nil {
			d.Accounts = append(d.Accounts, struct {
				Account journal.Account
				Empty   bool
			}{
				Account: a,
				Empty:   j.Balances[a].Empty(),
			})
		} else {
			d.Disabled = append(d.Disabled, a)
		}
	}
	return d
}

func makeTrialData(t *reports.TrialBalance) templates.TrialData {
	var rs []templates.TrialRow
	for _, tr := range t.Rows {
		r := templates.TrialRow{Account: string(tr.Account)}
		for _, p := range tr.Pairs {
			r.DebitBal = p.Debit
			r.CreditBal = p.Credit
			rs = append(rs, r)
			r = templates.TrialRow{}
		}
	}
	r := templates.TrialRow{Account: "Total"}
	for _, u := range balanceUnits(t.Total.Debit, t.Total.Credit) {
		r.DebitBal = t.Total.Debit.Amount(u)
		r.CreditBal = t.Total.Credit.Amount(u)
		rs = append(rs, r)
		r = templates.TrialRow{}
	}
	return templates.TrialData{Rows: rs}
}

func makeLedgerData(l *reports.AccountLedger) templates.LedgerData {
	d := templates.LedgerData{Account: l.Account}
	lastRef := ""
	for _, r := range l.Rows {
		r2 := templates.LedgerRow{
			Pair: r.Pair,
		}
		// The ledger template groups based on the date field.
		// We deduplicate splits/rows from the same
		// transaction based on the ref.
		if r.Ref != lastRef {
			r2.Date = r.Date.String()
			r2.Description = r.Description
			r2.Ref = r.Ref
			lastRef = r.Ref
		}
		if r.Pair.Debit != nil {
			r2.Balance = r.Balance.Amount(r.Pair.Debit.Unit)
		} else if r.Pair.Credit != nil {
			r2.Balance = r.Balance.Amount(r.Pair.Credit.Unit)
		} else if u := r.Balance.Units(); len(u) == 1 {
			r2.Balance = r.Balance.Amount(u[0])
		}
		d.Rows = append(d.Rows, r2)
	}
	return d
}

func makeStmtRows(a []journal.Account, b journal.Balances) ([]templates.StmtRow, *journal.Balance) {
	t := new(journal.Balance)
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

// balanceUnits returns all of the units in the balances.
func balanceUnits(b ...journal.Balance) []journal.Unit {
	seen := make(map[journal.Unit]bool)
	for _, b := range b {
		for _, u := range b.Units() {
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

// A stmt helps construct StmtData and add rows.
type stmt struct {
	*templates.StmtData
	// Tracks the running total
	bal journal.Balance
}

func (s *stmt) addRows(r ...templates.StmtRow) {
	s.Rows = append(s.Rows, r...)
}

// Adds a stmtRow with a balance.
// Since balances have multiple units, this adds multiple rows.
// The argument row's fields, except for the amount, is used for the
// first row only.
func (s *stmt) addBalanceRows(r templates.StmtRow, b *journal.Balance) {
	for _, v := range b.Amounts() {
		r.Amount = v
		s.addRows(r)
		r = templates.StmtRow{}
	}
}

func (s *stmt) addSection(desc string) {
	s.addRows(templates.StmtRow{
		Description: desc,
		Section:     true,
	})
}

// Adds rows for an account.
// The account balance is added to a running total.
func (s *stmt) addAccount(a journal.Account, b *journal.Balance) {
	s.addBalanceRows(templates.StmtRow{
		Description: string(a),
		Account:     true,
	}, b)
	s.bal.AddBal(b)
}

// Like addAccount but with amount.
func (s *stmt) addAccountAmount(a journal.Account, am *journal.Amount) {
	s.addRows(templates.StmtRow{
		Description: string(a),
		Amount:      am,
		Account:     true,
	})
	s.bal.Add(am)
}

// Add the current balance as a total.
func (s *stmt) addTotal(desc string) {
	s.addBalanceRows(templates.StmtRow{Description: desc}, &s.bal)
	s.bal.Clear()
}

// accountDeltas returns the accumulated balance changes of all
// transactions that affect the given accounts.
func accountDeltas(e []journal.Entry, a []journal.Account, start civil.Date) journal.Balances {
	e = filterEntries(e, newAccountPred(a).match)
	delta := make(journal.Balances)
	for _, e := range e {
		t, ok := e.(*journal.Transaction)
		if !ok {
			continue
		}
		if e.Date().Before(start) {
			continue
		}
		for _, s := range t.Splits {
			delta.Add(s.Account, s.Amount)
		}
	}
	return delta
}

// Filter entries related to the given account.
func accountEntries(e []journal.Entry, a journal.Account) []journal.Entry {
	return filterEntries(e, func(a2 journal.Account) bool {
		return a2 == a
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

// An accountPred can be used in filterEntries to match entries
// matching any listed account.
type accountPred map[journal.Account]bool

func newAccountPred(a []journal.Account) accountPred {
	p := make(accountPred)
	for _, a := range a {
		p[a] = true
	}
	return p
}

func (p accountPred) match(a journal.Account) bool {
	return p[a]
}

func cashAccounts(c *config.Config, a []journal.Account) []journal.Account {
	var new []journal.Account
	for _, a := range a {
		if c.IsCash(a) {
			new = append(new, a)
		}
	}
	return new
}

func sortedAccounts(j *journal.Journal) []journal.Account {
	var new []journal.Account
	for a := range j.Accounts {
		new = append(new, a)
	}
	sort.Slice(new, func(i, j int) bool { return new[i] < new[j] })
	return new
}
