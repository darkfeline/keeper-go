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

package book

import (
	"fmt"
	"testing"
)

func newAmount(n int64, symbol string, scale int64) Amount {
	return Amount{
		Number: n,
		UnitType: &UnitType{
			Symbol: symbol,
			Scale:  scale,
		},
	}
}

func TestAmount_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		a    Amount
		want string
	}{
		{"unit", newAmount(1234, "JPY", 1), "1234 JPY"},
		{"fractions", newAmount(12345, "USD", 100), "123.45 USD"},
		{"negative fractions", newAmount(-12345, "USD", 100), "-123.45 USD"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := c.a.String()
			if got != c.want {
				t.Errorf("Got %#v, want %#v", got, c.want)
			}
		})
	}
}

func ExampleWalkAccountTree() {
	f := func(n AccountNode) error {
		fmt.Println(n)
		return nil
	}
	a := []Account{
		"Equity:InitialBalance",
		"Assets:Cash",
		"Liabilities:CreditCard",
		"Income:Salary",
		"Expenses:Food",
		"Expenses:Hobbies:Games",
		"Expenses:Hobbies:Music",
	}
	_ = WalkAccountTree(a, f)
	// Output:
	// {Assets true}
	// {Assets:Cash false}
	// {Equity true}
	// {Equity:InitialBalance false}
	// {Expenses true}
	// {Expenses:Food false}
	// {Expenses:Hobbies true}
	// {Expenses:Hobbies:Games false}
	// {Expenses:Hobbies:Music false}
	// {Income true}
	// {Income:Salary false}
	// {Liabilities true}
	// {Liabilities:CreditCard false}
}
