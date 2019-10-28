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

package parse

import (
	"sort"
	"strings"

	"go.felesatra.moe/keeper/book"
)

type acctBalance []book.Amount

func (b *acctBalance) Add(a book.Amount) {
	if a.Number == 0 {
		return
	}
	for i := range *b {
		if (*b)[i].UnitType == a.UnitType {
			(*b)[i].Number += a.Number
			return
		}
	}
	*b = append(*b, a)
}

func (b acctBalance) Equal(b2 acctBalance) bool {
	c, c2 := b.copy(), b2.copy()
	c.removeEmpty()
	c.sort()
	c2.removeEmpty()
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

func (b *acctBalance) removeEmpty() {
	var new acctBalance
	for _, a := range *b {
		if a.Number != 0 {
			new = append(new, a)
		}
	}
	*b = new
}

func (b acctBalance) copy() acctBalance {
	new := make(acctBalance, len(b))
	copy(new, b)
	return new
}

func (b acctBalance) sort() {
	sort.Slice(b, func(i, j int) bool { return b[i].UnitType.Symbol < b[j].UnitType.Symbol })
}

func (b acctBalance) String() string {
	n := len(b)
	if n == 0 {
		return "empty"
	}
	s := make([]string, n)
	for i, a := range b {
		s[i] = a.String()
	}
	return strings.Join(s, ", ")
}
