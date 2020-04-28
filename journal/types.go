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

package journal

import (
	"sort"
	"strings"
)

// Account is a bookkeeping account.
// Accounts are colon separated paths, like "Income:Salary".
type Account string

// Parts returns the parts of the account between the colons.
func (a Account) Parts() []string {
	if a == "" {
		return nil
	}
	return strings.Split(string(a), ":")
}

// Parent returns the parent account.
func (a Account) Parent() Account {
	p := a.Parts()
	if len(p) == 0 {
		return ""
	}
	return Account(strings.Join(p[:len(p)-1], ":"))
}

// Leaf returns the leaf part of the Account (after the last colon).
func (a Account) Leaf() string {
	if a == "" {
		return ""
	}
	p := a.Parts()
	return p[len(p)-1]
}

// Under returns true if the Account is a child account of the argument.
func (a Account) Under(parent Account) bool {
	if parent == "" {
		return true
	}
	return strings.HasPrefix(string(a), string(parent)+":")
}

// An Amount is an amount of a certain unit, e.g., currency or commodity.
type Amount struct {
	Number int64
	Unit   Unit
}

// Neg returns the amount with its sign negated.
func (a Amount) Neg() Amount {
	a.Number = -a.Number
	return a
}

// Scalar returns the amount number without the unit as a formatted string.
func (a Amount) Scalar() string {
	return decFormat(a.Number, a.Unit.Scale)
}

func (a Amount) String() string {
	if (a.Unit == Unit{}) {
		return decFormat(a.Number, 1)
	}
	return decFormat(a.Number, a.Unit.Scale) + " " + a.Unit.Symbol
}

// Unit describes a unit, e.g., currency or commodity.
type Unit struct {
	// Symbol for the unit.
	// This should be all uppercase ASCII letters.
	Symbol string
	// Scale indicates the minimum fractional unit amount,
	// e.g. 100 means 0.01 is the smallest amount.
	// This should be a multiple of 10.
	Scale int64
}

func (u Unit) String() string {
	return u.Symbol
}

// A Balance represents a balance of amounts of various units.
type Balance map[Unit]int64

// Add adds an amount to the balance.
func (b Balance) Add(a Amount) {
	b[a.Unit] += a.Number
}

// Sub subtracts an amount from the balance.
func (b Balance) Sub(a Amount) {
	b.Add(a.Neg())
}

// Neg negates the sign of the balance.
func (b Balance) Neg() {
	for u, n := range b {
		b[u] = -n
	}
}

// Empty returns true if the balance is empty/zero.
func (b Balance) Empty() bool {
	for _, n := range b {
		if n != 0 {
			return false
		}
	}
	return true
}

// Amount returns the amount of the given unit in the balance.
func (b Balance) Amount(u Unit) Amount {
	return Amount{Number: b[u], Unit: u}
}

// Amounts returns the amounts in the balance.
// The amounts are sorted by unit.
func (b Balance) Amounts() []Amount {
	var a []Amount
	for u, n := range b {
		if n != 0 {
			a = append(a, Amount{Unit: u, Number: n})
		}
	}
	sort.Slice(a, func(i, j int) bool { return a[i].Unit.Symbol < a[j].Unit.Symbol })
	return a
}

// Equal returns true if the two balances are equal.
func (b Balance) Equal(b2 Balance) bool {
	b = b.Copy()
	for _, a := range b2.Amounts() {
		b.Sub(a)
	}
	return b.Empty()
}

// Copy returns a copy of the balance.
func (b Balance) Copy() Balance {
	new := make(Balance)
	for u, n := range b {
		if n != 0 {
			new[u] = n
		}
	}
	return new
}

func (b Balance) String() string {
	if len(b) == 0 {
		return "0"
	}
	s := make([]string, len(b))
	for i, a := range b.Amounts() {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}

// A Balances maps multiple accounts to their balances.
type Balances map[Account]Balance

// Add adds an amount to an account, even if the account is not yet in
// the map.
func (b Balances) Add(a Account, am Amount) {
	bal, ok := b[a]
	if !ok {
		bal = make(Balance)
		b[a] = bal
	}
	bal.Add(am)
}

// Neg negates the signs of the balances.
func (b Balances) Neg() {
	for _, b := range b {
		b.Neg()
	}
}

// A Summary tracks the total balance including sub-accounts for all accounts.
type Summary map[Account]Balance

// Add adds an amount to an account.
func (s Summary) Add(a Account, am Amount) {
	for a != "" {
		s.add1(a, am)
		a = a.Parent()
	}
}

// Adds an amount to an account, even if the account is not yet in
// the map.
func (s Summary) add1(a Account, am Amount) {
	bal, ok := s[a]
	if !ok {
		bal = make(Balance)
		s[a] = bal
	}
	bal.Add(am)
}

// Neg negates the signs of the balances.
func (s Summary) Neg() {
	for _, s := range s {
		s.Neg()
	}
}
