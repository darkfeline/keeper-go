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
	"go.felesatra.moe/keeper"
)

type Balance map[AccountType]TypeBalance
type TypeBalance map[string][]keeper.Quantity

func MakeBalance(ts []Transaction) Balance {
	b := make(Balance)
	b[Assets] = make(TypeBalance)
	b[Liabilities] = make(TypeBalance)
	b[Equity] = make(TypeBalance)
	b[Revenues] = make(TypeBalance)
	b[Expenses] = make(TypeBalance)
	for _, t := range ts {
		b.change(t.From, t.Amount.Neg(), t.Unit)
		b.change(t.To, t.Amount, t.Unit)
	}
	return b
}

func (b Balance) change(a Account, f keeper.Fixed, c keeper.Unit) {
	m, ok := b[a.Type]
	if !ok {
		panic(a.Type)
	}
	m[a.Name] = keeper.AddUnits(m[a.Name], f, c)
}
