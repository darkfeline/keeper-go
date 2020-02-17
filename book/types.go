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

package book

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Amount is an amount of a certain unit, e.g., currency or commodity.
type Amount struct {
	Number int64
	Unit   Unit
}

// Scalar returns the amount number without the unit as a formatted string.
func (a Amount) Scalar() string {
	return decFormat(a.Number, a.Unit.Scale)
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", decFormat(a.Number, a.Unit.Scale), a.Unit.Symbol)
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
	return fmt.Sprintf("%v (1/%v)", u.Symbol, u.Scale)
}

// Account is a bookkeeping account.
// Accounts are colon separated paths, like "Income:Salary".
type Account string

// Parts returns the parts of the account between the colons.
// An empty slice is returned for the zero value.
func (a Account) Parts() []string {
	if a == "" {
		return nil
	}
	return strings.Split(string(a), ":")
}

// Level returns the nesting level of the account, which is equivalent
// to the number of parts.
func (a Account) Level() int {
	return len(a.Parts())
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
	p := a.Parts()
	return p[len(p)-1]
}

// Under returns true if the Account is a child account of the argument.
func (a Account) Under(parent Account) bool {
	return strings.HasPrefix(string(a), string(parent)+":")
}

// Balance represents a balance of amounts of various units.
// The order of different units does not matter.
// There should not be more than one Amount for a unit.
type Balance []Amount

// Add adds an amount to the balance.
func (b Balance) Add(a Amount) Balance {
	if a.Number == 0 {
		return b
	}
	for i := range b {
		if b[i].Unit == a.Unit {
			b[i].Number += a.Number
			return b
		}
	}
	b = append(b, a)
	return b
}

// Empty returns true if the balance is empty/zero.
func (b Balance) Empty() bool {
	return len(b.CleanCopy()) == 0
}

// Equal returns true if the two balances are equal.
func (b Balance) Equal(b2 Balance) bool {
	c, c2 := b.CleanCopy(), b2.CleanCopy()
	if len(c) != len(c2) {
		return false
	}
	c.sort()
	c2.sort()
	return reflect.DeepEqual(c, c2)
}

// CleanCopy returns a copy of the balance without units that have
// zero amounts.
func (b Balance) CleanCopy() Balance {
	var new Balance
	for _, a := range b {
		if a.Number != 0 {
			new = append(new, a)
		}
	}
	return new
}

func (b Balance) sort() {
	sort.Slice(b, func(i, j int) bool { return b[i].Unit.Symbol < b[j].Unit.Symbol })
}

func (b Balance) String() string {
	n := len(b)
	if n == 0 {
		return "0"
	}
	s := make([]string, n)
	for i, a := range b {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}

// TBalance is a "trial balance", containing the balances of multiple accounts.
type TBalance map[Account]Balance
