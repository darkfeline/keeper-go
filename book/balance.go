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

package book

import (
	"sort"
	"strings"
)

// Balance represents a balance of amounts of various unit types.
// The order of different units does not matter.
// There should not be more than one Amount for a unit type.
// Equality of unit types is by pointer.
type Balance []Amount

// Add adds an amount to the balance.
func (b Balance) Add(a Amount) Balance {
	if a.Number == 0 {
		return b
	}
	for i := range b {
		if b[i].UnitType == a.UnitType {
			b[i].Number += a.Number
			return b
		}
	}
	b = append(b, a)
	return b
}

// Equal returns true if the two balances are equal.
func (b Balance) Equal(b2 Balance) bool {
	c, c2 := b.CleanCopy(), b2.CleanCopy()
	c.sort()
	c2.sort()
	if len(c) != len(c2) {
		return false
	}
	for i, a := range c {
		if a != c2[i] {
			return false
		}
	}
	return true
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
	sort.Slice(b, func(i, j int) bool { return b[i].UnitType.Symbol < b[j].UnitType.Symbol })
}

func (b Balance) String() string {
	n := len(b)
	if n == 0 {
		return "empty balance"
	}
	s := make([]string, n)
	for i, a := range b {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}
