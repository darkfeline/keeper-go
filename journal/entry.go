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

package journal

import (
	"cloud.google.com/go/civil"

	"go.felesatra.moe/keeper/kpr/token"
)

type Entry interface {
	Date() civil.Date
	Position() token.Position
	entry()
}

type BalanceAssert struct {
	EntryPos  token.Position
	EntryDate civil.Date
	Account   Account
	Balance   Balance
}

func (b BalanceAssert) Position() token.Position {
	return b.EntryPos
}

func (b BalanceAssert) Date() civil.Date {
	return b.EntryDate
}

func (BalanceAssert) entry() {}

// Transaction describes a bookkeeping transaction.
// The sum of all split amounts for all unit types should be zero.
type Transaction struct {
	EntryPos    token.Position
	EntryDate   civil.Date
	Description string
	Splits      []Split
}

func (t Transaction) Position() token.Position {
	return t.EntryPos
}

func (t Transaction) Date() civil.Date {
	return t.EntryDate
}

func (Transaction) entry() {}

// Split is one split in a transaction.
// This describes a change in the amount of one unit for one account.
type Split struct {
	Account Account
	Amount  Amount
}
