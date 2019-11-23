// Copyright (C) 2019  Allen Li
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

package parse

import (
	"errors"
	"fmt"
	"io"

	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/raw"
)

// Result contains the parse results.
type Result struct {
	Lines  []interface{}
	Errors []error
}

// Transactions returns the parsed transactions.
// This doesn't take into account any errors during parsing.
func (r Result) Transactions() []book.Transaction {
	var t []book.Transaction
	for _, l := range r.Lines {
		e, ok := l.(TransactionLine)
		if !ok {
			continue
		}
		t = append(t, e.Transaction())
	}
	return t
}

// Parse parses keeper format entries.
// See the module description for the format.
func Parse(r io.Reader) (Result, error) {
	entries, err := raw.Parse(r)
	if err != nil {
		return Result{}, fmt.Errorf("keeper parse: %v", err)
	}
	sortEntries(entries)
	p := newProcessor()
	for _, e := range entries {
		p.processEntry(e)
	}
	return Result{
		Lines:  p.lines,
		Errors: p.errors,
	}, nil
}

type processor struct {
	units    map[string]*book.UnitType
	balances map[book.Account]book.Balance

	transactions []book.Transaction
	lines        []interface{}
	errors       []error
}

func newProcessor() *processor {
	return &processor{
		units:    make(map[string]*book.UnitType),
		balances: make(map[book.Account]book.Balance),
	}
}

func (p *processor) processEntry(e interface{}) {
	var err error
	switch e := e.(type) {
	case raw.UnitEntry:
		err = p.processUnit(e)
	case raw.BalanceEntry:
		err = p.processBalance(e)
	case raw.TransactionEntry:
		err = p.processTransaction(e)
	default:
		panic(fmt.Sprintf("unknown entry: %#v", e))
	}
	if err != nil {
		p.errors = append(p.errors, err)
	}
}

func (p *processor) processUnit(u raw.UnitEntry) error {
	if _, ok := p.units[u.Symbol]; ok {
		return processErrf(u, "symbol %v already declared", u.Symbol)
	}
	scale, err := decimalToInt64(u.Scale)
	if err != nil {
		return processErr(u, err)
	}
	if scale < 0 {
		return processErrf(u, "negative scale")
	}
	if !isPower10(scale) {
		return processErrf(u, "scale")
	}
	p.units[u.Symbol] = &book.UnitType{
		Symbol: u.Symbol,
		Scale:  scale,
	}
	return nil
}

func (p *processor) processBalance(b raw.BalanceEntry) error {
	want, err := p.convertBalances(b.Amounts)
	if err != nil {
		return processErr(b, err)
	}
	got := p.balances[b.Account]
	l := BalanceLine{
		Common: Common{
			Date: b.Date,
			Line: b.Line,
		},
		Account:  b.Account,
		Balance:  got,
		Declared: want,
	}
	if !got.Equal(want) {
		l.Err = fmt.Errorf("balance %v not equal to declared %v", got, want)
		err = processErr(b, l.Err)
	}
	p.lines = append(p.lines, l)
	return err
}

func (p *processor) processTransaction(t raw.TransactionEntry) error {
	l := TransactionLine{
		Common: Common{
			Date: t.Date,
			Line: t.Line,
		},
		Description: t.Description,
	}
	var err error
	if l.Splits, err = p.convertSplits(t.Splits); err != nil {
		return processErr(t, err)
	}
	if err := fillEmptySplit(l.Splits); err != nil {
		return processErr(t, err)
	}
	// Transaction is now good enough to keep even if it has other errors.
	for _, s := range l.Splits {
		p.addToBalance(s)
	}
	if b, _ := splitsBalance(l.Splits); len(b) > 0 {
		l.Err = fmt.Errorf("unbalanced amount %v", b)
		err = processErr(t, l.Err)
	}
	p.lines = append(p.lines, l)
	return err
}

// addToBalance updates the running balance with the split.
func (p *processor) addToBalance(s book.Split) {
	b := p.balances[s.Account]
	p.balances[s.Account] = b.Add(s.Amount)
}

func (p *processor) convertSplits(s []raw.Split) ([]book.Split, error) {
	bs := make([]book.Split, len(s))
	for i, s := range s {
		s2, err := p.convertSplit(s)
		if err != nil {
			return nil, err
		}
		bs[i] = s2
	}
	return bs, nil
}

func (p *processor) convertSplit(s raw.Split) (book.Split, error) {
	var a book.Amount
	if s.Amount != (raw.Amount{}) {
		var err error
		a, err = p.convertAmount(s.Amount)
		if err != nil {
			return book.Split{}, fmt.Errorf("convert split %v: %v", s.Account, err)
		}
	}
	return book.Split{
		Account: s.Account,
		Amount:  a,
	}, nil
}

func (p *processor) convertBalances(a []raw.Amount) (book.Balance, error) {
	var b book.Balance
	for _, a := range a {
		a2, err := p.convertAmount(a)
		if err != nil {
			return nil, err
		}
		b = append(b, a2)
	}
	return b, nil
}

func (p *processor) convertAmount(a raw.Amount) (book.Amount, error) {
	u, ok := p.units[a.Unit]
	if !ok {
		return book.Amount{}, fmt.Errorf("convert amount %v: unknown unit", a)
	}
	a2, err := combineDecimalUnit(a.Number, u)
	if err != nil {
		return book.Amount{}, fmt.Errorf("convert amount %v: %v", a, err)
	}
	return a2, nil
}

func fillEmptySplit(s []book.Split) error {
	b, empty := splitsBalance(s)
	switch len(empty) {
	case 0:
	case 1:
		if len(b) != 1 {
			return fmt.Errorf("fill empty split: unsuitable balance %v", b)
		}
		(*(empty[0])).Amount = b[0].Neg()
	default:
		return errors.New("fill empty split: multiple empty")
	}
	return nil
}

// splitsBalance returns the balance for the splits.
func splitsBalance(s []book.Split) (b book.Balance, empty []*book.Split) {
	for i := range s {
		if s[i].Amount == (book.Amount{}) {
			empty = append(empty, &s[i])
			continue
		}
		b = b.Add(s[i].Amount)
	}
	return b.CleanCopy(), empty
}

// combineDecimalUnit combines a decimal magnitude and unit into a book amount.
func combineDecimalUnit(d raw.Decimal, u *book.UnitType) (book.Amount, error) {
	if d.Scale > u.Scale {
		rescale := d.Scale / u.Scale
		if d.Number%rescale != 0 {
			return book.Amount{}, fmt.Errorf("%v fractions too small for unit %v", d, u)
		}
		d.Number /= rescale
		d.Scale /= rescale
	}
	return book.Amount{
		Number:   d.Number * u.Scale / d.Scale,
		UnitType: u,
	}, nil
}

func decimalToInt64(d raw.Decimal) (int64, error) {
	if d.Number%d.Scale != 0 {
		return 0, fmt.Errorf("decimal to int64 %v: non-integer", d)
	}
	return d.Number / d.Scale, nil
}

func isPower10(n int64) bool {
	if n < 0 {
		n = -n
	}
	if n == 1 {
		return true
	}
	for x := int64(10); x <= n; x *= 10 {
		if x == n {
			return true
		}
	}
	return false
}
