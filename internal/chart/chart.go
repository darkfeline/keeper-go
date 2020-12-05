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

// Package chart is not stable.
package chart

import (
	"fmt"
	"io"
	"strings"

	"github.com/pelletier/go-toml"
	"go.felesatra.moe/keeper/journal"
)

type Account = journal.Account

type Config struct {
	CashAcctPrefix []string `toml:"cash_account_prefix"`
}

func LoadConfig(c *Config, r io.Reader) error {
	d := toml.NewDecoder(r)
	if err := d.Decode(c); err != nil {
		return fmt.Errorf("load chart config: %s", err)
	}
	return nil
}

func (c *Config) IsIncome(a Account) bool {
	return a.Under("Income")
}

func (c *Config) IsExpenses(a Account) bool {
	return a.Under("Expenses")
}

func (c *Config) IsAssets(a Account) bool {
	return a.Under("Assets")
}

func (c *Config) IsLiabilities(a Account) bool {
	return a.Under("Liabilities")
}

func (c *Config) IsEquity(a Account) bool {
	return a.Under("Equity")
}

func (c *Config) IsTrading(a Account) bool {
	return a.Under("Trading")
}

func (c *Config) IsCash(a Account) bool {
	for _, p := range c.CashAcctPrefix {
		if strings.HasPrefix(string(a), p) {
			return true
		}
	}
	return false
}
