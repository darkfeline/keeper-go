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
	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

type BalanceEntry struct {
	Date    civil.Date
	Account book.Account
	Amounts []Amount
}

type Amount struct {
	Number Decimal
	Unit   string
}

type UnitEntry struct {
	Symbol string
	Scale  Decimal
}

type TransactionEntry struct {
	Date        civil.Date
	Description string
	Splits      []Split
}

type Split struct {
	Account book.Account
	Amount  Amount
}