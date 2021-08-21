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

package journal

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/parser"
	"go.felesatra.moe/keeper/kpr/scanner"
	"go.felesatra.moe/keeper/kpr/token"
)

// buildEntries builds entries from keeper file source.
// This is done in a single pass on an entry by entry basis, so
// balances are not tracked.
// Each transaction must still balance to zero however.
func buildEntries(inputs ...inputBytes) ([]Entry, error) {
	fset := token.NewFileSet()
	b := newBuilder(fset)
	var entries []Entry
	for _, i := range inputs {
		f, err := parser.ParseBytes(fset, i.filename, i.src, 0)
		if err != nil {
			return nil, fmt.Errorf("build entries: %s", err)
		}
		e, err := b.build(f.Entries)
		if err != nil {
			return entries, fmt.Errorf("build entries: %s", err)
		}
		entries = append(entries, e...)
	}
	return entries, nil
}

type builder struct {
	fset  *token.FileSet
	units map[string]Unit
	errs  scanner.ErrorList
}

func newBuilder(fset *token.FileSet) *builder {
	return &builder{
		fset:  fset,
		units: make(map[string]Unit),
	}
}

func (b *builder) build(t []ast.Entry) ([]Entry, error) {
	var entries []Entry
	for _, n := range t {
		switch n := n.(type) {
		case *ast.SingleBalance:
			e, err := b.buildSingleBalance(n)
			if err != nil {
				continue
			}
			entries = append(entries, e)
		case *ast.MultiBalance:
			e, err := b.buildMultiBalance(n)
			if err != nil {
				continue
			}
			entries = append(entries, e)
		case *ast.Transaction:
			e, err := b.buildTransaction(n)
			if err != nil {
				continue
			}
			entries = append(entries, e)
		case *ast.UnitDecl:
			b.addUnit(n)
		case *ast.DisableAccount:
			e, err := b.buildDisableAccount(n)
			if err != nil {
				continue
			}
			entries = append(entries, e)
		default:
			panic(fmt.Sprintf("unknown entry node %T", n))
		}
	}
	return entries, b.errs.Err()
}

func (b *builder) nodePos(e ast.Node) token.Position {
	return b.fset.Position(e.Pos())
}

func (b *builder) errorf(pos token.Pos, format string, v ...interface{}) {
	b.errs.Add(b.fset.Position(pos), fmt.Sprintf(format, v...))
}

func (b *builder) buildSingleBalance(n *ast.SingleBalance) (*BalanceAssert, error) {
	a, err := b.buildBalanceHeader(&n.BalanceHeader)
	if err != nil {
		return a, err
	}
	amount, err := b.buildAmount(n.Amount)
	if err != nil {
		return a, err
	}
	a.Declared.Add(amount)
	return a, nil
}

func (b *builder) buildMultiBalance(n *ast.MultiBalance) (*BalanceAssert, error) {
	e, err := b.buildBalanceHeader(&n.BalanceHeader)
	if err != nil {
		return e, err
	}
	for _, n := range n.Amounts {
		line := n.(*ast.AmountLine)
		amount, err := b.buildAmount(line.Amount)
		if err != nil {
			return e, err
		}
		e.Declared.Add(amount)
	}
	return e, nil
}

func (b *builder) buildBalanceHeader(n *ast.BalanceHeader) (*BalanceAssert, error) {
	assertKind(n.Date, token.DATE)
	assertKind(n.Account, token.ACCTNAME)

	a := &BalanceAssert{
		EntryPos: b.nodePos(n),
		Account:  Account(n.Account.Value),
		Declared: make(Balance),
		// Actual and Diff get set when the entries are added
		// to the Journal, so we don't have to initialize them
		// now.
	}
	switch n.Token {
	case token.BALANCE:
	case token.TREEBAL:
		a.Tree = true
	default:
		panic(fmt.Sprintf("unexpected token %s", n.Token))
	}
	var err error
	a.EntryDate, err = civil.ParseDate(n.Date.Value)
	if err != nil {
		b.errorf(n.Date.Pos(), "%s", err)
		return a, err
	}

	return a, nil
}

func (b *builder) buildTransaction(n *ast.Transaction) (*Transaction, error) {
	assertKind(n.Date, token.DATE)
	assertKind(n.Description, token.STRING)

	t := &Transaction{
		EntryPos:    b.nodePos(n),
		Description: parseString(n.Description.Value),
	}
	var err error
	t.EntryDate, err = civil.ParseDate(n.Date.Value)
	if err != nil {
		b.errorf(n.Date.Pos(), "%s", err)
		return t, err
	}

	var empty *Split
	bal := make(Balance)
	t.Splits = make([]Split, len(n.Splits))
	for i, n := range n.Splits {
		n := n.(*ast.SplitLine)
		assertKind(n.Account, token.ACCTNAME)
		s := &t.Splits[i]
		s.Account = Account(n.Account.Value)
		if n.Amount == nil {
			if empty != nil {
				b.errorf(n.Pos(), "more than one split missing amount")
				return t, fmt.Errorf("more than one split missing amount")
			}
			empty = s
			continue
		}
		a, err := b.buildAmount(n.Amount)
		if err != nil {
			return t, err
		}
		s.Amount = a
		bal.Add(a)
	}
	switch empty {
	case nil:
		if !bal.Empty() {
			b.errorf(n.Pos(), "transaction doesn't balance (off by %s)", bal)
			return t, fmt.Errorf("transaction doesn't balance (off by %s)", bal)
		}
	default:
		amounts := bal.Amounts()
		if len(amounts) != 1 {
			b.errorf(n.Pos(), "cannot infer missing split amount with balance %s", bal)
			return t, fmt.Errorf("cannot infer missing split amount with balance %s", bal)
		}
		a := amounts[0].Neg()
		empty.Amount = a
	}
	return t, nil
}

func (b *builder) buildAmount(n *ast.Amount) (Amount, error) {
	assertKind(n.Decimal, token.DECIMAL)
	assertKind(n.Unit, token.USYMBOL)

	s := n.Decimal.Value
	s = strings.Replace(s, ",", "", -1)
	r := newRat()
	defer ratPool.Put(r)
	_, err := fmt.Sscan(s, r)
	if err != nil {
		b.errorf(n.Unit.Pos(), "%s", err)
		return Amount{}, err
	}

	sym := n.Unit.Value
	if !validateUnit(sym) {
		b.errorf(n.Unit.Pos(), "bad unit %s", sym)
		return Amount{}, fmt.Errorf("bad unit %s", sym)
	}
	u, ok := b.units[sym]
	if !ok {
		b.errorf(n.Unit.Pos(), "undeclared unit %s", sym)
		return Amount{}, fmt.Errorf("undeclared unit %s", sym)
	}

	r2 := newRat()
	defer ratPool.Put(r2)
	r.Mul(r, r2.SetInt64(u.Scale))
	if !r.IsInt() {
		b.errorf(n.Unit.Pos(), "scaled unit amount is fractional")
		return Amount{}, fmt.Errorf("scaled unit amount is fractional")
	}

	num, ok := numberFromInt(r.Num())
	if !ok {
		b.errorf(n.Unit.Pos(), "%q too big", n.Decimal.Value)
		return Amount{}, fmt.Errorf("%q too big", n.Decimal.Value)
	}
	a := Amount{
		Number: num,
		Unit:   u,
	}
	return a, nil
}

func (b *builder) buildDisableAccount(n *ast.DisableAccount) (*DisableAccount, error) {
	assertKind(n.Date, token.DATE)
	assertKind(n.Account, token.ACCTNAME)
	e := &DisableAccount{
		EntryPos: b.nodePos(n),
		Account:  Account(n.Account.Value),
	}
	var err error
	e.EntryDate, err = civil.ParseDate(n.Date.Value)
	if err != nil {
		b.errorf(n.Date.Pos(), "%s", err)
		return e, err
	}
	return e, nil
}

func (b *builder) addUnit(n *ast.UnitDecl) {
	assertKind(n.Unit, token.USYMBOL)
	assertKind(n.Scale, token.DECIMAL)

	s := n.Scale.Value
	s = strings.Replace(s, ",", "", -1)
	scale, err := strconv.ParseInt(s, 10, 64)
	switch {
	case err != nil:
		b.errorf(n.Scale.Pos(), "%s", err)
		return
	case scale < 0:
		b.errorf(n.Scale.Pos(), "negative scale")
		return
	case !isPower10(scale):
		b.errorf(n.Scale.Pos(), "scale not power of 10")
		return
	}

	unit := n.Unit.Value
	if !validateUnit(unit) {
		b.errorf(n.Unit.Pos(), "bad unit %s", unit)
		return
	}
	u := Unit{
		Symbol: unit,
		Scale:  scale,
	}
	if prev, ok := b.units[unit]; ok && prev != u {
		b.errorf(n.Unit.Pos(), "unit %s redeclared with different scale", unit)
		return
	}
	b.units[unit] = u
}

func validateUnit(lit string) (ok bool) {
	for _, r := range lit {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func assertKind(n *ast.BasicValue, tok token.Token) {
	if n.Kind != tok {
		panic(fmt.Sprintf("token was %s not %s", n.Kind, tok))
	}
}

func parseString(src string) string {
	if src[0] != '"' || src[len(src)-1] != '"' {
		panic(fmt.Sprintf("bad string %#v", src))
	}
	src = src[1 : len(src)-1]
	var out []rune
	var escape bool
	for _, r := range src {
		if escape {
			out = append(out, r)
			escape = false
			continue
		}
		if r == '\\' {
			escape = true
			continue
		}
		out = append(out, r)
	}
	return string(out)
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
