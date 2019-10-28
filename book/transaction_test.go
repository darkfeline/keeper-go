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
)

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
