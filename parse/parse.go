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

type Parser struct {
	entries  []interface{}
	units    map[string]*book.UnitType
	balances map[book.Account]book.Balance
	lines    []interface{}
	errors   []error
}

// Parse parses keeper format entries.
// See the module description for the format.
func Parse(r io.Reader) (*Parser, error) {
	entries, err := raw.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("keeper parse: %v", err)
	}
	sortEntries(entries)
	p := &Parser{
		entries:  entries,
		units:    make(map[string]*book.UnitType),
		balances: make(map[book.Account]book.Balance),
	}
	return p, nil
}

// CheckedTransactions returns parsed transactions.
// Only the final transactions are returned.
// The transactions are sorted by date and checked for correctness.
func (p *Parser) CheckedTransactions() ([]book.Transaction, error) {
	ps := newProcessor()
	var errs []error
process:
	for _, e := range p.entries {
		if len(errs) >= 20 {
			errs = append(errs, errors.New("(too many errors)"))
			break process
		}
		if err := ps.processEntry(e); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return ps.transactions, processError{errs}
	}
	return ps.transactions, nil
}

type processor struct {
	units        map[string]*book.UnitType
	balances     map[book.Account]book.Balance
	transactions []book.Transaction
}

func newProcessor() *processor {
	return &processor{
		units:    make(map[string]*book.UnitType),
		balances: make(map[book.Account]book.Balance),
	}
}

func (p *processor) processEntry(e interface{}) error {
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
	}
	var err error
	t2.Splits, err = p.convertSplits(t.Splits)
	if err != nil {
		return processErr(t, err)
	}

	b, empty := splitsBalance(t2.Splits)
	switch len(empty) {
	case 0:
	case 1:
		b = b.CleanCopy()
		if len(b) != 1 {
			return processErrf(t, "unsuitable balance for empty split %v", b)
		}
		(*(empty[0])).Amount = b[0].Neg()
		b = nil
	default:
		return processErrf(t, "multiple empty splits")
	}

	if len(b) > 0 {
		return processErrf(t, "unbalanced amount %v", b)
	}
	p.transactions = append(p.transactions, t2)
	for _, s := range t2.Splits {
		p.addToBalance(s)
	}
	return nil
}

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
	return convertAmount(a.Number, u)
}

func splitsBalance(s []book.Split) (b book.Balance, empty []*book.Split) {
	for i := range s {
		if s[i].Amount == (book.Amount{}) {
			empty = append(empty, &s[i])
			continue
		}
		b = b.Add(s[i].Amount)
	}
	return b, empty
}

func decimalToInt64(d raw.Decimal) (int64, error) {
	if d.Number%d.Scale != 0 {
		return 0, fmt.Errorf("decimal to int64 %v: non-integer", d)
	}
	return d.Number / d.Scale, nil
}

func convertAmount(d raw.Decimal, u *book.UnitType) (book.Amount, error) {
	if d.Scale > u.Scale {
		rescale := d.Scale / u.Scale
		if d.Number%rescale != 0 {
			return book.Amount{}, fmt.Errorf("convert amount: fractions for %v too small for unit %v", d, u)
		}
		d.Number /= rescale
		d.Scale /= rescale
	}
	return book.Amount{
		Number:   d.Number * u.Scale / d.Scale,
		UnitType: u,
	}, nil
}
