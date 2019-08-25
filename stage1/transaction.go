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
	"fmt"
	"strings"

	"go.felesatra.moe/keeper"
)

type Transaction struct {
	From   Account
	To     Account
	Amount keeper.Fixed
	Unit   keeper.Unit
}

type Account struct {
	Type AccountType
	Name string
}

func ParseAccount(s string) (Account, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Account{}, fmt.Errorf("parse account %#v: invalid", s)
	}
	t, ok := types[parts[0]]
	if !ok {
		return Account{}, fmt.Errorf("parse account %#v: invalid account type %v", s, parts[0])
	}
	return Account{
		Type: t,
		Name: parts[1],
	}, nil
}

func (a Account) String() string {
	return fmt.Sprintf("%s:%s", a.Type, a.Name)
}

type AccountType int

//go:generate stringer -type=AccountType

const (
	Assets AccountType = iota
	Liabilities
	Equity
	Revenues
	Expenses
)

var types = map[string]AccountType{
	"Assets":      Assets,
	"Liabilities": Liabilities,
	"Equity":      Equity,
	"Revenues":    Revenues,
	"Expenses":    Expenses,
}
