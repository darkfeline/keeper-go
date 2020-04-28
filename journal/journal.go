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

// Package journal implements bookkeeping journals with keeper files.
package journal

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// A Journal represents accounting information compiled from keeper file source.
type Journal struct {
	// Entries are all of the entries, sorted chronologically.
	Entries []Entry
	// Closed contains closed accounts.
	Closed map[Account]CloseAccount
	// Balances is the final balance for all accounts.
	Balances Balances
	// Summary is the total balance including sub-accounts for all accounts.
	Summary Summary
	// BalanceErrors contains the balance assertion entries that failed.
	BalanceErrors []BalanceAssert
}

// Compile compiles keeper file source into a Journal.
// Balance assertion errors are not returned here, to enable the
// caller to inspect the transactions to identify the error.
func Compile(o ...Option) (*Journal, error) {
	opts := makeOptions(o)
	inputs, err := openInputFiles(opts.inputs)
	e, err := buildEntries(inputs...)
	if err != nil {
		return nil, fmt.Errorf("keeper: %s", err)
	}
	sortEntries(e)
	if d := opts.ending; d.IsValid() {
		e = entriesEnding(e, d)
	}
	j, err := compile(e)
	if err != nil {
		return nil, fmt.Errorf("keeper: %s", err)
	}
	return j, nil
}

// openInputFiles reads inputFiles and replaces them with their contents.
func openInputFiles(inputs []input) ([]inputBytes, error) {
	var ib []inputBytes
	for _, i := range inputs {
		switch i := i.(type) {
		case inputBytes:
			ib = append(ib, i)
		case inputFile:
			src, err := ioutil.ReadFile(i.filename)
			if err != nil {
				return nil, err
			}
			ib = append(ib, inputBytes{
				filename: filepath.Base(i.filename),
				src:      src,
			})
		default:
			panic(fmt.Sprintf("unknown type %T", i))
		}
	}
	return ib, nil
}

// compile compiles a Journal from entries.
// Entries should be sorted.
func compile(e []Entry) (*Journal, error) {
	j := &Journal{
		Closed:   make(map[Account]CloseAccount),
		Balances: make(Balances),
		Summary:  make(Summary),
	}
	for _, e := range e {
		if err := j.addEntry(e); err != nil {
			return nil, err
		}
	}
	return j, nil
}

func (j *Journal) addEntry(e Entry) error {
	switch e := e.(type) {
	case Transaction:
		return j.addTransaction(e)
	case BalanceAssert:
		return j.addBalanceAssert(e)
	case CloseAccount:
		j.Entries = append(j.Entries, e)
		if err := j.checkAccountClosed(e.Account); err != nil {
			return fmt.Errorf("add entry %T at %s: %s", e, e.Position(), err)
		}
		j.Closed[e.Account] = e
		return nil
	default:
		panic(fmt.Sprintf("unknown Entry type %T", e))
	}
}

func (j *Journal) addTransaction(e Transaction) error {
	e.Balances = make(Balances)
	for _, s := range e.Splits {
		if err := j.checkAccountClosed(s.Account); err != nil {
			return fmt.Errorf("add entry %T at %s: %s", e, e.Position(), err)
		}
		j.Balances.Add(s.Account, s.Amount)
		j.Summary.Add(s.Account, s.Amount)
		e.Balances[s.Account] = j.Balances[s.Account].Copy()
	}
	j.Entries = append(j.Entries, e)
	return nil
}

func (j *Journal) addBalanceAssert(e BalanceAssert) error {
	if err := j.checkAccountClosed(e.Account); err != nil {
		return fmt.Errorf("add entry %T at %s: %s", e, e.Position(), err)
	}
	var m map[Account]Balance
	if e.Tree {
		m = j.Summary
	} else {
		m = j.Balances
	}
	bal, ok := m[e.Account]
	switch ok {
	case true:
		bal = bal.Copy()
	case false:
		bal = make(Balance)
	}
	e.Actual = bal
	e.Diff = balanceDiff(e.Actual, e.Declared)

	j.Entries = append(j.Entries, e)
	if !e.Diff.Empty() {
		j.BalanceErrors = append(j.BalanceErrors, e)
	}
	return nil
}

func (j *Journal) checkAccountClosed(a Account) error {
	if _, ok := j.Closed[a]; ok {
		return fmt.Errorf("account %s is closed", a)
	}
	return nil
}

func balanceDiff(x, y Balance) Balance {
	diff := x.Copy()
	for _, a := range y.Amounts() {
		diff.Sub(a)
	}
	return diff
}
