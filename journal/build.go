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
	"unicode"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/parser"
	"go.felesatra.moe/keeper/kpr/scanner"
	"go.felesatra.moe/keeper/kpr/token"
)

func Build(src []byte) ([]Entry, error) {
	fset := token.NewFileSet()
	t, err := parser.ParseBytes(fset, "", src, 0)
	if err != nil {
		return nil, fmt.Errorf("build journal: %s", err)
	}
	b := newBuilder(fset)
	entries, err := b.build(t)
	if err != nil {
		return entries, fmt.Errorf("build journal: %s", err)
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
		case ast.SingleBalance:
			b, err := b.buildSingleBalance(n)
			if err != nil {
				continue
			}
			entries = append(entries, b)
		case ast.MultiBalance:
			b, err := b.buildMultiBalance(n)
			if err != nil {
				continue
			}
			entries = append(entries, b)
		case ast.Transaction:
			b, err := b.buildTransaction(n)
			if err != nil {
				continue
			}
			entries = append(entries, b)
		case ast.UnitDecl:
			b.addUnit(n)
		default:
			panic(fmt.Sprintf("unknown entry node %T", n))
		}
	}
	if len(b.errs) > 0 {
		return entries, b.errs
	}
	return entries, nil
}

func (b *builder) errorf(pos token.Pos, format string, v ...interface{}) {
	b.errs.Add(b.fset.Position(pos), fmt.Sprintf(format, v...))
}

func (b *builder) buildSingleBalance(n ast.SingleBalance) (BalanceAssert, error) {
	a, err := b.buildBalanceHeader(n.BalanceHeader)
	if err != nil {
		return a, err
	}
	amount, err := b.buildAmount(n.Amount)
	if err != nil {
		return a, err
	}
	a.Balance = a.Balance.Add(amount)
	return a, nil
}

func (b *builder) buildMultiBalance(n ast.MultiBalance) (BalanceAssert, error) {
	e, err := b.buildBalanceHeader(n.BalanceHeader)
	if err != nil {
		return e, err
	}
	for _, n := range n.Amounts {
		line := n.(ast.AmountLine)
		amount, err := b.buildAmount(line.Amount)
		if err != nil {
			return e, err
		}
		e.Balance = e.Balance.Add(amount)
	}
	return e, nil
}

func (b *builder) buildBalanceHeader(n ast.BalanceHeader) (BalanceAssert, error) {
	assertKind(n.Date, token.DATE)
	assertKind(n.Account, token.ACCOUNT)

	var a BalanceAssert
	var err error
	a.EntryDate, err = civil.ParseDate(n.Date.Value)
	if err != nil {
		b.errorf(n.Date.Pos(), "%s", err)
		return a, err
	}

	a.Account = Account(n.Account.Value)
	return a, nil
}

func (b *builder) buildTransaction(n ast.Transaction) (Transaction, error) {
	assertKind(n.Date, token.DATE)
	assertKind(n.Description, token.STRING)

	var t Transaction
	var err error
	t.EntryDate, err = civil.ParseDate(n.Date.Value)
	if err != nil {
		b.errorf(n.Date.Pos(), "%s", err)
		return t, err
	}
	t.Description = parseString(n.Description.Value)

	var empty *Split
	var bal Balance
	for i, n := range n.Splits {
		n := n.(ast.Split)
		assertKind(n.Account, token.ACCOUNT)
		t.Splits = append(t.Splits, Split{
			Account: Account(n.Account.Value),
		})
		s := &t.Splits[i]
		if n.Amount == nil {
			if empty != nil {
				b.errorf(n.Pos(), "more than one split missing amount")
				return t, fmt.Errorf("more than one split missing amount")
			}
			empty = s
			continue
		}
		a, err := b.buildAmount(*n.Amount)
		if err != nil {
			return t, err
		}
		s.Amount = a
		bal = bal.Add(a)
	}
	bal = bal.CleanCopy()
	if empty != nil {
		if len(bal) != 1 {
			b.errorf(n.Pos(), "cannot infer missing split amount with balance %s", bal)
			return t, fmt.Errorf("cannot infer missing split amount with balance %s", bal)
		}
		a := bal[0]
		a.Number = -a.Number
		empty.Amount = a
		bal = bal.Add(a).CleanCopy()
	}
	if len(bal) != 0 {
		b.errorf(n.Pos(), "transaction doesn't balance (off by %s)", bal)
		return t, fmt.Errorf("transaction doesn't balance (off by %s)", bal)
	}
	return t, nil
}

func (b *builder) buildAmount(n ast.Amount) (Amount, error) {
	assertKind(n.Decimal, token.DECIMAL)
	assertKind(n.Unit, token.IDENT)

	d, err := parseDecimal(n.Decimal.Value)
	if err != nil {
		b.errorf(n.Decimal.Pos(), "%s", err)
		return Amount{}, err
	}

	sym := n.Unit.Value
	if !validateUnit(sym) {
		b.errorf(n.Unit.Pos(), "bad unit %s", sym)
		return Amount{}, err
	}
	u, ok := b.units[sym]
	if !ok {
		b.errorf(n.Unit.Pos(), "undeclared unit %s", sym)
		return Amount{}, fmt.Errorf("undeclared unit %s", sym)
	}

	a, err := combineDecimalUnit(d, u)
	if err != nil {
		b.errorf(n.Pos(), "%s", err)
		return Amount{}, err
	}
	return a, nil
}

func (b *builder) addUnit(n ast.UnitDecl) {
	assertKind(n.Unit, token.IDENT)
	assertKind(n.Scale, token.DECIMAL)

	d, err := parseDecimal(n.Scale.Value)
	if err != nil {
		b.errorf(n.Scale.Pos(), "%s", err)
		return
	}
	scale, err := decimalToInt64(d)
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
	b.units[unit] = Unit{
		Symbol: unit,
		Scale:  scale,
	}
}

func validateUnit(lit string) (ok bool) {
	for _, r := range lit {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func assertKind(n ast.BasicValue, tok token.Token) {
	if n.Kind != tok {
		panic(fmt.Sprintf("token was %s not %s", n.Kind, tok))
	}
}

// combineDecimalUnit combines a decimal magnitude and unit into an amount.
func combineDecimalUnit(d decimal, u Unit) (Amount, error) {
	if d.scale > u.Scale {
		rescale := d.scale / u.Scale
		if d.number%rescale != 0 {
			return Amount{}, fmt.Errorf("%v fractions too small for unit %v", d, u)
		}
		d.number /= rescale
		d.scale /= rescale
	}
	return Amount{
		Number: d.number * u.Scale / d.scale,
		Unit:   u,
	}, nil
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

func decimalToInt64(d decimal) (int64, error) {
	if d.number%d.scale != 0 {
		return 0, fmt.Errorf("decimal to int64 %v: non-integer", d)
	}
	return d.number / d.scale, nil
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
