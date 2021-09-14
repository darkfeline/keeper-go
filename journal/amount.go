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

// Zero returns whether the amount is zero.
func (a *Amount) Zero() bool {
	return isZero(&a.Number)
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
// The zero value is ready for use.
type Balance struct {
	m map[Unit]*big.Int
}

// Gets the Int for a Unit, initializing it if needed.
func (b *Balance) get(u Unit) *big.Int {
	if b.m == nil {
		b.m = make(map[Unit]*big.Int)
	}
	n := b.m[u]
	if n == nil {
		n = newInt()
		b.m[u] = n
	}
	return n
}

// Add adds an amount to the balance.
func (b *Balance) Add(a *Amount) {
	n := b.get(a.Unit)
	n.Add(n, &a.Number)
}

// AddBal adds the amounts of the argument balance.
func (b *Balance) AddBal(b2 *Balance) {
	if b2 == nil {
		return
	}
	for k, v := range b2.m {
		if isZero(v) {
			continue
		}
		n := b.get(k)
		n.Add(n, v)
	}
}

// Neg negates the sign of the balance.
func (b *Balance) Neg() {
	for _, v := range b.m {
		v.Neg(v)
	}
}

// Empty returns true if the balance is empty/zero.
func (b *Balance) Empty() bool {
	if b == nil {
		return true
	}
	for _, v := range b.m {
		if !isZero(v) {
			return false
		}
	}
	return true
}

// Clear clears the balance, making it empty/zero.
func (b *Balance) Clear() {
	for k := range b.m {
		delete(b.m, k)
	}
}

// Has returns true if the balance has a non-zero amount for the unit.
func (b *Balance) Has(u Unit) bool {
	n := b.m[u]
	return n != nil && !isZero(n)
}

// Amount returns the amount of the given unit in the balance.
func (b *Balance) Amount(u Unit) *Amount {
	a := &Amount{Unit: u}
	if n := b.m[u]; n != nil && !isZero(n) {
		a.Number.Set(n)
	}
	return a
}

// Units returns the units in the balance
func (b *Balance) Units() []Unit {
	var us []Unit
	for k, v := range b.m {
		if isZero(v) {
			continue
		}
		us = append(us, k)
	}
	return us
}

// Amounts returns the amounts in the balance.
// The amounts are sorted by unit.
func (b *Balance) Amounts() []*Amount {
	var as []*Amount
	for k, v := range b.m {
		if isZero(v) {
			continue
		}
		a := &Amount{Unit: k}
		a.Number.Set(v)
		as = append(as, a)
	}
	sort.Slice(as, func(i, j int) bool {
		return as[i].Unit.Symbol < as[j].Unit.Symbol
	})
	return as
}

// Equal returns true if the two balances are equal.
func (b *Balance) Equal(b2 *Balance) bool {
	var b3 Balance
	b3.Set(b)
	b3.Neg()
	b3.AddBal(b2)
	return b3.Empty()
}

// Set sets the receiver balance to the argument balance.
func (b *Balance) Set(b2 *Balance) {
	b.Clear()
	if b2 == nil {
		return
	}
	for k, v := range b2.m {
		if !isZero(v) {
			b.get(k).Set(v)
		}
	}
}

func (b *Balance) String() string {
	amts := b.Amounts()
	if len(amts) == 0 {
		return "0"
	}
	s := make([]string, len(amts))
	for i, a := range amts {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}

// A Balances maps multiple accounts to their balances.
type Balances map[Account]*Balance

// Add adds an amount to an account, even if the account is not yet in
// the map.
func (b Balances) Add(a Account, am *Amount) {
	bal, ok := b[a]
	if !ok {
		bal = new(Balance)
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
