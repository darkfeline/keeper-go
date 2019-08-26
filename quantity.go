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

package keeper

import (
	"fmt"
)

type Quantity struct {
	Amount Decimal
	Unit   Unit
}

type Unit string

func NewQuantity(value int64, point uint8, u Unit) Quantity {
	return Quantity{
		Amount: NewDecimal(value, point),
		Unit:   u,
	}
}

func (q Quantity) Neg() Quantity {
	q.Amount = q.Amount.Neg()
	return q
}

func (q Quantity) String() string {
	return fmt.Sprintf("%v %v", q.Amount, q.Unit)
}

func (q *Quantity) Increase(f Decimal) {
	q.Amount = q.Amount.Add(f)
}

func AddQuantity(qs []Quantity, q Quantity) []Quantity {
	for i, _ := range qs {
		if qs[i].Unit == q.Unit {
			qs[i].Increase(q.Amount)
			return qs
		}
	}
	return append(qs, q)
}

func MergeQuantities(q1 []Quantity, q2 []Quantity) []Quantity {
	for _, q := range q2 {
		q1 = AddQuantity(q1, q)
	}
	return q1
}
