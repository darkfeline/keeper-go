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

/*
Package journal implements the processing of bookkeeping journals from
keeper files.

# File format

See the documentation for the kpr package for keeper file syntax.
This section describes some additional semantics for keeper files as
implemented by this package.

Unit declarations must come before any use of that unit.  Otherwise,
the order of entries in keeper files is not significant.

The total of all splits in a transaction must balance.  Only one split
in a transaction can omit the amount, which will be inferred as the
remaining amount needed to balance the transaction.  This inference
does not work if more than one unit is unbalanced.

Balance assertions apply at the end of the day, to match how balances
are handled in practice.

Tree balance assertions apply to a tree of accounts.

Disabled accounts prevent transactions from posting to that account.
Disable account entries also assert that the account balance is zero.
*/
package journal

import (
	"fmt"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/parser"
	"go.felesatra.moe/keeper/kpr/token"
)

// Compile compiles keeper file source into a Journal.
// Balance assertion errors are stored in the returned Journal rather
// than returned as errors here, to enable the caller to inspect the
// transactions to identify the error.
func Compile(a *CompileArgs) (*Journal, error) {
	// Compiling a journal happens in stages:
	//  1. Parse inputs into ast entries
	//  2. Convert ast entries into journal entries ("building")
	//  3. Sort entries by date
	//  4. Go through entries adding up balances and checking things ("compiling")
	//  5. Fill in account metadata
	fset := token.NewFileSet()
	e, err := parseEntries(fset, a.Inputs...)
	if err != nil {
		return nil, fmt.Errorf("compile journal: %s", err)
	}
	b := newBuilder(fset)
	e2, err := b.build(e...)
	if err != nil {
		return nil, fmt.Errorf("compile journal: %s", err)
	}
	sortEntries(e2)
	if d := a.Ending; d.IsValid() {
		e2 = entriesEnding(e2, d)
	}
	j, err := compile(e2)
	if err != nil {
		return nil, fmt.Errorf("compile journal: %s", err)
	}
	copyAccountMetadata(b, j)
	return j, nil
}

// parseEntries parses inputs into ast entries.
func parseEntries(fset *token.FileSet, inputs ...CompileInput) ([]ast.Entry, error) {
	var e []ast.Entry
	for _, i := range inputs {
		src, err := i.Src()
		if err != nil {
			return nil, fmt.Errorf("build entries: %s", err)
		}
		f, err := parser.ParseBytes(fset, i.Filename(), src, 0)
		if err != nil {
			return nil, fmt.Errorf("build entries: %s", err)
		}
		e = append(e, f.Entries...)
	}
	return e, nil
}

// compile compiles a Journal from entries.
// Entries should be sorted.
func compile(e []Entry) (*Journal, error) {
	j := newJournal()
	for _, e := range e {
		if err := j.addEntry(e); err != nil {
			return nil, err
		}
	}
	return j, nil
}

// copyAccountMetadata copies account metadata from the builder to the
// journal.
func copyAccountMetadata(b *builder, j *Journal) {
	for a, ai := range b.accounts {
		ai2, ok := j.Accounts[a]
		if !ok {
			// TODO(darkfeline): We allocate a map for
			// metadata that we immediately throw away
			// below.
			ai2 = newAccountInfo()
			j.Accounts[a] = ai2
		}
		ai2.Metadata = ai.Metadata
	}
}

// A Journal represents bookkeeping information compiled from keeper file source.
//
// Be careful; a Journal contains a lot of shared pointers internally.
// Modifying anything in a Journal is not recommended.
type Journal struct {
	// Entries are the journal entries, sorted chronologically.
	Entries []Entry
	// Accounts contains all accounts and the associated account information.
	Accounts AccountMap
	// Balances is the final balance for all accounts.
	Balances Balances
	// BalanceErrors contains the balance assertion entries that failed.
	BalanceErrors []*BalanceAssert
}

// newJournal makes a new Journal.
func newJournal() *Journal {
	return &Journal{
		Accounts: make(AccountMap),
		Balances: make(Balances),
	}
}

// BalancesEnding returns the balances of all accounts at the close of
// the given date.
func (j *Journal) BalancesEnding(d civil.Date) Balances {
	b := make(Balances)
	for _, e := range j.Entries {
		t, ok := e.(*Transaction)
		if !ok {
			continue
		}
		if e.Date().After(d) {
			break
		}
		for _, s := range t.Splits {
			b.Add(s.Account, s.Amount)
		}
	}
	return b
}

func (j *Journal) addEntry(e Entry) error {
	switch e := e.(type) {
	case *Transaction:
		return j.addTransaction(e)
	case *BalanceAssert:
		return j.addBalanceAssert(e)
	case *DisableAccount:
		return j.addDisableAccount(e)
	default:
		panic(fmt.Sprintf("unknown Entry type %T", e))
	}
}

func (j *Journal) addTransaction(e *Transaction) error {
	for _, s := range e.Splits {
		j.ensureAccount(s.Account)
		if err := j.checkAccountDisabled(s.Account); err != nil {
			return fmt.Errorf("add entry %T at %s: %s", e, e.Position(), err)
		}
		j.Balances.Add(s.Account, s.Amount)
	}
	j.Entries = append(j.Entries, e)
	return nil
}

func (j *Journal) addBalanceAssert(e *BalanceAssert) error {
	j.ensureAccount(e.Account)
	if err := j.checkAccountDisabled(e.Account); err != nil {
		return fmt.Errorf("add entry %T at %s: %s", e, e.Position(), err)
	}
	if e.Tree {
		addTreeBalance(&e.Actual, j.Balances, e.Account)
	} else {
		e.Actual.Set(j.Balances[e.Account])
	}
	e.Diff.Set(&e.Declared)
	e.Diff.Neg()
	e.Diff.AddBal(&e.Actual)

	j.Entries = append(j.Entries, e)
	if !e.Diff.Empty() {
		j.BalanceErrors = append(j.BalanceErrors, e)
	}
	return nil
}

func (j *Journal) addDisableAccount(e *DisableAccount) error {
	j.ensureAccount(e.Account)
	if err := j.checkAccountDisabled(e.Account); err != nil {
		return fmt.Errorf("add entry %T at %s: %s", e, e.Position(), err)
	}
	if bal := j.Balances[e.Account]; bal != nil && !bal.Empty() {
		ba := &BalanceAssert{
			EntryPos:  e.EntryPos,
			EntryDate: e.EntryDate,
			Account:   e.Account,
		}
		ba.Actual.Set(bal)
		ba.Diff.Set(bal)
		j.BalanceErrors = append(j.BalanceErrors, ba)
	}
	j.Accounts[e.Account].Disabled = e
	j.Entries = append(j.Entries, e)
	return nil
}

func (j *Journal) checkAccountDisabled(a Account) error {
	if e := j.Accounts[a].Disabled; e != nil {
		return fmt.Errorf("account %s is disabled by entry %s", a, e)
	}
	return nil
}

// Ensure account is present in accounts map.
func (j *Journal) ensureAccount(a Account) {
	if j.Accounts[a] == nil {
		j.Accounts[a] = newAccountInfo()
	}
}

// addTreeBalance returns the total balance for an account and
// sub-accounts.
func addTreeBalance(b *Balance, bals Balances, a Account) {
	b.AddBal(bals[a])
	for a2, b2 := range bals {
		if a2.Under(a) {
			b.AddBal(b2)
		}
	}
}
