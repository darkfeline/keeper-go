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
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/parse/internal/raw"
)

func Parse(r io.Reader) ([]book.Transaction, error) {
	entries, err := raw.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("keeper parse: %v", err)
	}
	sortEntries(entries)
	p := newProcessor()
	for _, e := range entries {
		p.processEntry(e)
	}
	if len(p.errs) != 0 {
		return p.transactions, ProcessError{Errors: p.errs}
	}
	panic("Not implemented")
}

type ProcessError struct {
	Errors []error
}

func (e ProcessError) Error() string {
	n := len(e.Errors)
	if n == 0 {
		return "error while processing"
	}
	s := make([]string, n)
	for i, e := range e.Errors {
		s[i] = e.Error()
	}
	return fmt.Sprintf("%d errors while processing:\n  -%v",
		len(e.Errors),
		strings.Join(s, "\n  -"))
}

type processor struct {
	units        map[string]book.UnitType
	transactions []book.Transaction
	errs         []error
}

func newProcessor() *processor {
	return &processor{
		units: make(map[string]book.UnitType),
	}
}

func (p *processor) processEntry(e interface{}) {}

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

func convertAmount(d raw.Decimal, u book.UnitType) (book.Amount, error) {
	if d.Scale > u.Scale {
		return book.Amount{}, fmt.Errorf("amount %v for unit %v divisions too small", d, u)
	}
	return book.Amount{
		Number:   d.Number * u.Scale / d.Scale,
		UnitType: u,
	}, nil
}
