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

import "go.felesatra.moe/keeper/book"

type acctBalance []book.Amount

func (b *acctBalance) Add(a book.Amount) {
	for _, a2 := range *b {
		if a2.UnitType == a.UnitType {
			a2.Number += a.Number
			return
		}
	}
	*b = append(*b, a)
}
