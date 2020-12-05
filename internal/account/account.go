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

// Package account contains account related utilities.
package account

import (
	"fmt"
	"io"
	"strings"

	"github.com/pelletier/go-toml"
	"go.felesatra.moe/keeper/journal"
)

type Account = journal.Account

type Classifier struct {
	// Prefix for matching cash accounts.
	CashPrefix []string `toml:"cash_prefix"`
}

func LoadClassifier(c *Classifier, r io.Reader) error {
	d := toml.NewDecoder(r)
	if err := d.Decode(c); err != nil {
		return fmt.Errorf("load account classifier: %s", err)
	}
	return nil
}

func (c *Classifier) IsIncome(a Account) bool {
	return a.Under("Income")
}

func (c *Classifier) IsExpenses(a Account) bool {
	return a.Under("Expenses")
}

func (c *Classifier) IsAssets(a Account) bool {
	return a.Under("Assets")
}

func (c *Classifier) IsLiabilities(a Account) bool {
	return a.Under("Liabilities")
}

func (c *Classifier) IsEquity(a Account) bool {
	return a.Under("Equity")
}

func (c *Classifier) IsTrading(a Account) bool {
	return a.Under("Trading")
}

func (c *Classifier) IsCash(a Account) bool {
	for _, p := range c.CashPrefix {
		if strings.HasPrefix(string(a), p) {
			return true
		}
	}
	return false
}
