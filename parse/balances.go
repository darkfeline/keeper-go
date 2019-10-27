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
	for i := range *b {
		if (*b)[i].UnitType == a.UnitType {
			(*b)[i].Number += a.Number
			return
		}
	}
	*b = append(*b, a)
}

func (b acctBalance) Equal(b2 acctBalance) bool {
	b.Sort()
	b2.Sort()
	for i, a := range b {
		if a != b2[i] {
			return false
		}
	}
	return true
}

func (b acctBalance) Sort() {
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
