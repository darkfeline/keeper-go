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

import "go.felesatra.moe/keeper/journal"

type Account = journal.Account

type Chart struct {
	income      []Account
	expenses    []Account
	assets      []Account
	liabilities []Account
	equity      []Account
}

func New(a []Account) *Chart {
	c := &Chart{}
	for _, a := range a {
		switch {
		case a.Under("Income"):
			c.income = append(c.income, a)
		case a.Under("Expenses"):
			c.expenses = append(c.expenses, a)
		case a.Under("Assets"):
			c.assets = append(c.assets, a)
		case a.Under("Liabilities"):
			c.liabilities = append(c.liabilities, a)
		case a.Under("Equity"):
			c.equity = append(c.equity, a)
		}
	}
	return c
}

func (c *Chart) Income() []Account {
	return c.income
}

func (c *Chart) Expenses() []Account {
	return c.expenses
}

func (c *Chart) Assets() []Account {
	return c.assets
}

func (c *Chart) Liabilities() []Account {
	return c.liabilities
}

func (c *Chart) Equity() []Account {
	return c.equity
}
