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
)

type Transaction struct {
	Date        civil.Date
	Description string
	Splits      []Split
}

type Split struct {
	Account Account
	Amount  Amount
}

type Amount struct {
	Number   int64
	UnitType UnitType
}

func (a Amount) String() string {
	u := a.UnitType
	if u.Scale <= 1 {
		return fmt.Sprintf("%d %s", a.Number, u.Symbol)
	}
	return fmt.Sprintf("%d.%d %s", a.Number/u.Scale, a.Number%u.Scale, u.Symbol)
}

type UnitType struct {
	Symbol string
	// Scale indicates the minimum fractional unit amount,
	// e.g. 100 means 0.01 is the smallest amount.
	Scale int64
}

func (u UnitType) String() string {
	return fmt.Sprintf("%v (1/%v)", u.Symbol, u.Scale)
}
