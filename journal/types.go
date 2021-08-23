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
	"sync"
)

// A Number is n[0] + n[1] * (1 << 63)
type Number [2]int64

const numWidth = 1 << 63

// Neg returns the number with its sign negated.
func (n Number) Neg() Number {
	for i := range n {
		n[i] = -n[i]
	}
	return n
}

// IsNeg returns true if the number is negative.
func (n Number) IsNeg() bool {
	for i := range n {
		if n[i] < 0 {
			return true
		}
	}
	return false
}

// Zero returns true if the number is zero.
func (n Number) Zero() bool {
	for i := range n {
		if n[i] != 0 {
			return false
		}
	}
	return true
}

func (n Number) setRat(r *big.Rat) {
	r2 := newRat()
	defer ratPool.Put(r2)
	r.SetInt64(n[1])
	r.Mul(r, r2.SetUint64(numWidth))
	r.Add(r, r2.SetInt64(n[0]))
}

// modifies input
func numberFromInt(i *big.Int) (Number, bool) {
	i2 := newInt()
	defer intPool.Put(i2)
	i3 := newInt()
	defer intPool.Put(i3)
	i.QuoRem(i, i2.SetUint64(numWidth), i3)
	if !i3.IsInt64() {
		panic(i3)
	}
	if !i.IsInt64() {
		return Number{}, false
	}
	return Number{i3.Int64(), i.Int64()}, true
}

// An Amount is an amount of a certain unit, e.g., currency or commodity.
type Amount struct {
	Number Number
	Unit   Unit
}

// Neg returns the amount with its sign negated.
func (a Amount) Neg() Amount {
	a.Number = a.Number.Neg()
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
type Balance map[Unit]Number

// Add adds an amount to the balance.
func (b Balance) Add(a Amount) {
	v := b[a.Unit]
	for i := range v {
		v[i] += a.Number[i]
	}
	b[a.Unit] = v
}

// AddBal adds the amounts of the argument balance.
func (b Balance) AddBal(b2 Balance) {
	for k, v := range b2 {
		b.Add(Amount{
			Number: v,
			Unit:   k,
		})
	}
}

// Sub subtracts an amount from the balance.
func (b Balance) Sub(a Amount) {
	b.Add(a.Neg())
}

// Neg negates the sign of the balance.
func (b Balance) Neg() {
	for k, v := range b {
		b[k] = v.Neg()
	}
}

// Empty returns true if the balance is empty/zero.
func (b Balance) Empty() bool {
	for _, v := range b {
		if !v.Zero() {
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

// Amount returns the amount of the given unit in the balance.
func (b Balance) Amount(u Unit) Amount {
	return Amount{Number: b[u], Unit: u}
}

// Amounts returns the amounts in the balance.
// The amounts are sorted by unit.
func (b Balance) Amounts() []Amount {
	var as []Amount
	for k, v := range b {
		if !v.Zero() {
			as = append(as, Amount{Unit: k, Number: v})
		}
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
		if !v.Zero() {
			new[k] = v
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

// An AccountInfo holds account information.
type AccountInfo struct {
	// If the account is disabled, points to the entry that
	// disabled the account.  Otherwise this is nil.
	Disabled *DisableAccount
	Metadata map[string]string
}
