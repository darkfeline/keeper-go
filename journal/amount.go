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
	"math/big"
	"sort"
	"strings"
)

// An Amount is an amount of a certain unit, e.g., currency or commodity.
type Amount struct {
	Number big.Int
	Unit   Unit
}

// Neg flips the sign of the amount.
func (a *Amount) Neg() {
	a.Number.Neg(&a.Number)
}

// Equal returns true if the amounts are equal.
func (a *Amount) Equal(b *Amount) bool {
	return a.Unit == b.Unit && a.Number.Cmp(&b.Number) == 0
}

func (a *Amount) String() string {
	return decFormat(&a.Number, a.Unit.Scale) + " " + a.Unit.Symbol
}

// Unit describes a unit, e.g., currency or commodity.
type Unit struct {
	// Symbol for the unit.
	// This should be all uppercase ASCII letters.
	Symbol string
	// Scale indicates the minimum fractional unit amount,
	// e.g. 100 means 0.01 is the smallest amount.
	// This should be a multiple of 10.
	Scale uint64
}

func (u Unit) String() string {
	return u.Symbol
}

// A Balance represents a balance of amounts of various units.
type Balance map[Unit]*big.Int

// Gets the Int for a Unit, initializing it if needed.
func (b Balance) get(u Unit) *big.Int {
	n := b[u]
	if n == nil {
		n = newInt()
		b[u] = n
	}
	return n
}

// Add adds an amount to the balance.
func (b Balance) Add(a *Amount) {
	n := b.get(a.Unit)
	n.Add(n, &a.Number)
}

// AddBal adds the amounts of the argument balance.
func (b Balance) AddBal(b2 Balance) {
	for k, v := range b2 {
		n := b.get(k)
		n.Add(n, v)
	}
}

// Sub subtracts an amount from the balance.
func (b Balance) Sub(a *Amount) {
	n := b.get(a.Unit)
	n.Sub(n, &a.Number)
}

// Neg negates the sign of the balance.
func (b Balance) Neg() {
	for _, v := range b {
		v.Neg(v)
	}
}

// Empty returns true if the balance is empty/zero.
func (b Balance) Empty() bool {
	for _, v := range b {
		if !isZero(v) {
			return false
		}
	}
	return true
}

// Clear clears the balance, making it empty/zero.
func (b Balance) Clear() {
	for k := range b {
		delete(b, k)
	}
}

// Has returns true if the balance has a non-zero amount for the unit.
func (b Balance) Has(u Unit) bool {
	n := b[u]
	return n != nil && !isZero(n)
}

// Amount returns the amount of the given unit in the balance.
func (b Balance) Amount(u Unit) *Amount {
	a := &Amount{Unit: u}
	n := b[u]
	if n != nil {
		a.Number.Set(n)
	}
	return a
}

// Amounts returns the amounts in the balance.
// The amounts are sorted by unit.
func (b Balance) Amounts() []*Amount {
	var as []*Amount
	for k, v := range b {
		if isZero(v) {
			continue
		}
		as = append(as, b.Amount(k))
	}
	sort.Slice(as, func(i, j int) bool {
		return as[i].Unit.Symbol < as[j].Unit.Symbol
	})
	return as
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
// If called with a nil receiver, returns an empty initialized Balance.
func (b Balance) Copy() Balance {
	new := make(Balance)
	for k, v := range b {
		if !isZero(v) {
			new.get(k).Set(v)
		}
	}
	return new
}

func (b Balance) String() string {
	if len(b) == 0 {
		return "0"
	}
	amts := b.Amounts()
	s := make([]string, len(amts))
	for i, a := range amts {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}

// A Balances maps multiple accounts to their balances.
type Balances map[Account]Balance

// Add adds an amount to an account, even if the account is not yet in
// the map.
func (b Balances) Add(a Account, am *Amount) {
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

// Accounts returns all of the accounts with balances in sorted order.
func (b Balances) Accounts() []Account {
	var new []Account
	for a, b := range b {
		if b.Empty() {
			continue
		}
		new = append(new, a)
	}
	sort.Slice(new, func(i, j int) bool { return new[i] < new[j] })
	return new
}

func isZero(n *big.Int) bool {
	return len(n.Bits()) == 0
}
