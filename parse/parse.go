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
	for _, e := range entries {
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
		return fmt.Errorf("process unit: symbol %v already declared", u.Symbol)
	}
	scale, err := decimalToInt64(u.Scale)
	if err != nil {
		return fmt.Errorf("process unit: %v", err)
	}
	p.units[u.Symbol] = &book.UnitType{
		Symbol: u.Symbol,
		Scale:  scale,
	}
	return nil
}

func (p *processor) processBalance(u raw.BalanceEntry) error {
	panic("Not implemented")
}

func (p *processor) processTransaction(u raw.TransactionEntry) error {
	panic("Not implemented")
}

func (p *processor) convertAmount(a raw.Amount) (book.Amount, error) {
	u, ok := p.units[a.Unit]
	if !ok {
		return book.Amount{}, fmt.Errorf("convert amount %v: unknown unit", a)
	}
	return convertAmount(a.Number, u)
}

func sortEntries(e []interface{}) {
	type keyed struct {
		k int64
		v interface{}
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
