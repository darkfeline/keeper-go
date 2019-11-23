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
	"fmt"

	"cloud.google.com/go/civil"

	"go.felesatra.moe/keeper/internal/decfmt"
)

// Transaction describes a bookkeeping transaction.
// The sum of all split amounts for all unit types should be zero.
type Transaction struct {
	Date        civil.Date
	Description string
	Splits      []Split
}

// Split is one split in a transaction.  This describes a change in
// the amount of one unit type for one account.
type Split struct {
	Account Account
	Amount  Amount
}

// Amount is an amount of a certain unit, e.g., currency or commodity.
type Amount struct {
	// Number is the number of the smallest unit of the UnitType.
	Number   int64
	UnitType *UnitType
}

// Neg returns the additive inverse of the amount.
func (a Amount) Neg() Amount {
	a.Number = -a.Number
	return a
}

func (a Amount) Scalar() string {
	return decfmt.Format(a.Number, a.UnitType.Scale)
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.Scalar(), a.UnitType.Symbol)
}

// UnitType describes a unit, e.g., currency or commodity.
type UnitType struct {
	// Symbol for the unit.
	// This should be all uppercase ASCII letters.
	Symbol string
	// Scale indicates the minimum fractional unit amount,
	// e.g. 100 means 0.01 is the smallest amount.
	// This should be a multiple of 10.
	Scale int64
}

func (u UnitType) String() string {
	return fmt.Sprintf("%v (1/%v)", u.Symbol, u.Scale)
}
