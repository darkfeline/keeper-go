// Copyright (C) 2020  Allen Li
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

package config

import (
	"fmt"
	"io"
	"strings"

	"github.com/pelletier/go-toml"
	"go.felesatra.moe/keeper/journal"
)

type Config struct {
	Account `toml:"account"`
}

type Account struct {
	// Prefix for matching cash accounts.
	CashPrefix []string `toml:"cash_prefix"`
}

func Load(c *Config, r io.Reader) error {
	d := toml.NewDecoder(r)
	if err := d.Decode(c); err != nil {
		return fmt.Errorf("load account classifier: %s", err)
	}
	return nil
}

func (*Account) IsIncome(a journal.Account) bool {
	return a.Under("Income")
}

func (*Account) IsExpenses(a journal.Account) bool {
	return a.Under("Expenses")
}

func (*Account) IsAssets(a journal.Account) bool {
	return a.Under("Assets")
}

func (*Account) IsLiabilities(a journal.Account) bool {
	return a.Under("Liabilities")
}

func (*Account) IsEquity(a journal.Account) bool {
	return a.Under("Equity")
}

func (*Account) IsTrading(a journal.Account) bool {
	return a.Under("Trading")
}

func (c *Account) IsCash(a journal.Account) bool {
	for _, p := range c.CashPrefix {
		if strings.HasPrefix(string(a), p) {
			return true
		}
	}
	return false
}
