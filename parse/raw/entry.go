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
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/book"
)

// Entry describes the common interface implemented by all
// entries.  Type assertions can be used to process specific types of
// entries.
type Entry interface {
	// Line returns the line number of the entry.
	Line() int
	// Summary returns a string that usefully identifies the
	// entry, e.g., in errors.
	Summary() string
}

// common contains the fields common to all entries.
type common struct {
	line int
}

func (c common) Line() int {
	return c.line
}

// BalanceEntry represents a balance entry.
type BalanceEntry struct {
	common
	Date    civil.Date
	Account book.Account
	Amounts []Amount
}

func (e BalanceEntry) Summary() string {
	return fmt.Sprintf("balance %v %v", e.Date, e.Account)
}

func (e BalanceEntry) Equal(v BalanceEntry) bool {
	return e.common == v.common &&
		e.Date == v.Date &&
		e.Account == v.Account &&
		cmp.Equal(e.Amounts, v.Amounts)
}

// Amount represents an amount of a unit.
type Amount struct {
	Number Decimal
	Unit   string
}

func (a Amount) String() string {
	return fmt.Sprintf("%v %s", a.Number, a.Unit)
}

// UnitEntry represents a unit entry.
type UnitEntry struct {
	common
	Symbol string
	Scale  Decimal
}

func (e UnitEntry) Summary() string {
	return fmt.Sprintf("unit %v", e.Symbol)
}

func (e UnitEntry) Equal(v UnitEntry) bool {
	return e == v
}

// TransactionEntry represents a transaction entry.
type TransactionEntry struct {
	common
	Date        civil.Date
	Description string
	Splits      []Split
}

func (e TransactionEntry) Summary() string {
	return fmt.Sprintf("transaction %v %#v", e.Date, e.Description)
}

func (e TransactionEntry) Equal(v TransactionEntry) bool {
	return e.common == v.common &&
		e.Date == v.Date &&
		e.Description == v.Description &&
		cmp.Equal(e.Splits, v.Splits)
}

// Split represents one split in a transaction.
type Split struct {
	Account book.Account
	Amount  Amount
}
