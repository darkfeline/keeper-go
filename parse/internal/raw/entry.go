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

package raw

import (
	"fmt"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

type EntryCommon interface {
	EntryCommon() Common
	// Summary returns a string that usefully identifies the
	// entry, e.g., in errors.
	Summary() string
}

type Common struct {
	Line int
}

func (c Common) EntryCommon() Common {
	return c
}

type BalanceEntry struct {
	Common
	Date    civil.Date
	Account book.Account
	Amounts []Amount
}

func (e BalanceEntry) Summary() string {
	return fmt.Sprintf("balance %v %v", e.Date, e.Account)
}

type Amount struct {
	Number Decimal
	Unit   string
}

func (a Amount) String() string {
	return fmt.Sprintf("%v %s", a.Number, a.Unit)
}

type UnitEntry struct {
	Common
	Symbol string
	Scale  Decimal
}

func (e UnitEntry) Summary() string {
	return fmt.Sprintf("unit %v", e.Symbol)
}

type TransactionEntry struct {
	Common
	Date        civil.Date
	Description string
	Splits      []Split
}

func (e TransactionEntry) Summary() string {
	return fmt.Sprintf("transaction %v %#v", e.Date, e.Description)
}

type Split struct {
	Account book.Account
	Amount  Amount
}
