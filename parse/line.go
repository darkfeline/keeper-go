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

import (
	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

type Common struct {
	Date civil.Date
	Line int
}

type TransactionLine struct {
	Common
	Description string
	Splits      []book.Split
	Errors      []error
}

func (l TransactionLine) Transaction() book.Transaction {
	return book.Transaction{
		Date:        l.Date,
		Description: l.Description,
		Splits:      l.Splits,
	}
}

type BalanceLine struct {
	Common
	Account book.Account
	Amounts []book.Amount
	Errors  []error
}
