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

// Package parse implements parsing of keeper entries.
package parse

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/internal/raw"
)

// Parse parses keeper format entries.
// See the module description for the format.
// Only the final transactions are returned.
// The transactions are sorted by date and checked for correctness.
func Parse(r io.Reader) ([]book.Transaction, error) {
	entries, err := raw.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("keeper parse: %v", err)
	}
	sortEntries(entries)
	p := newProcessor()
	var errs []error
process:
	for _, e := range entries {
		if len(errs) >= 20 {
			errs = append(errs, errors.New("(too many errors)"))
			break process
		}
		if err := p.processEntry(e); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return p.transactions, processError{errs}
	}
	return p.transactions, nil
}

type processor struct {
	units        map[string]*book.UnitType
	balances     map[book.Account]acctBalance
	transactions []book.Transaction
}

func newProcessor() *processor {
	return &processor{
		units: make(map[string]*book.UnitType),
	}
}

func (p *processor) processEntry(e raw.EntryCommon) error {
	switch e := e.(type) {
	case raw.UnitEntry:
		return p.processUnit(e)
	case raw.BalanceEntry:
		return p.processBalance(e)
	case raw.TransactionEntry:
		return p.processTransaction(e)
	default:
		panic(fmt.Sprintf("unknown entry: %#v", e))
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
	if !got.Equal(want) {
		return processErrf(b, "balance %v not equal to declared %v", got, want)
	}
	return nil
}

func (p *processor) processTransaction(t raw.TransactionEntry) error {
	t2 := book.Transaction{
		Date:        t.Date,
		Description: t.Description,
		Splits:      make([]book.Split, len(t.Splits)),
	}
	var b acctBalance
	// emptySplit := -1
	for i, s := range t.Splits {
		s2, err := p.convertSplit(s)
		if err != nil {
			return processErr(t, err)
		}
		t2.Splits[i] = s2
		if s2.Amount == (book.Amount{}) {
			// if emptySplit
		} else {
			b.Add(s2.Amount)
		}
	}
	panic("Not implemented")
	return nil
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

func (p *processor) convertBalances(a []raw.Amount) (acctBalance, error) {
	var b acctBalance
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
	return convertAmount(a.Number, u)
}

func sortEntries(e []raw.EntryCommon) {
	type keyed struct {
		k int64
		v raw.EntryCommon
	}
	ks := make([]keyed, len(e))
	for i, e := range e {
		ks[i] = keyed{entryKey(e), e}
	}
	sort.Slice(ks, func(i, j int) bool {
		return ks[i].k < ks[j].k
	})
	for i, k := range ks {
		e[i] = k.v
	}
}

// entryKey returns a sort key corresponding to an entry.
func entryKey(e interface{}) int64 {
	switch e := e.(type) {
	case raw.TransactionEntry:
		return dateKey(e.Date)
	case raw.BalanceEntry:
		return dateKey(e.Date) + 1
	default:
		return 0
	}
}

// dateKey returns a sort key corresponding to a Date.
func dateKey(d civil.Date) int64 {
	return d.In(time.UTC).Unix()
}

func decimalToInt64(d raw.Decimal) (int64, error) {
	if d.Fraction() != 0 {
		return 0, fmt.Errorf("decimal to int64 %v: non-integer", d)
	}
	return d.Integer(), nil
}

func convertAmount(d raw.Decimal, u *book.UnitType) (book.Amount, error) {
	if d.Scale > u.Scale {
		return book.Amount{}, fmt.Errorf("amount %v for unit %v divisions too small", d, u)
	}
	return book.Amount{
		Number:   d.Number * u.Scale / d.Scale,
		UnitType: u,
	}, nil
}
