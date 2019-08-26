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

package stage1

import (
	"sort"

	"go.felesatra.moe/keeper"
)

type Balances map[keeper.Account][]keeper.Quantity

func MakeBalance(ts []Transaction) Balances {
	b := make(Balances)
	for _, t := range ts {
		b[t.From] = keeper.AddQuantity(b[t.From], t.Quantity.Neg())
		b[t.To] = keeper.AddQuantity(b[t.To], t.Quantity)
	}
	return b
}

func (b Balances) Accounts() []keeper.Account {
	as := make([]keeper.Account, 0, len(b))
	for a := range b {
		as = append(as, a)
	}
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
	return as
}
