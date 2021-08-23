// Copyright (C) 2021  Allen Li
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

import "fmt"

func ExampleAccount_Parts() {
	fmt.Printf("%#v\n", Account("Assets:Cash").Parts())
	fmt.Printf("%#v\n", Account("Assets").Parts())
	// Output:
	// []string{"Assets", "Cash"}
	// []string{"Assets"}
}

func ExampleAccount_Parent() {
	fmt.Printf("%#v\n", Account("Assets:Cash").Parent())
	fmt.Printf("%#v\n", Account("Assets").Parent())
	// Output:
	// "Assets"
	// ""
}

func ExampleAccount_Leaf() {
	fmt.Printf("%#v\n", Account("Assets:Cash").Leaf())
	fmt.Printf("%#v\n", Account("Assets").Leaf())
	// Output:
	// "Cash"
	// "Assets"
}

func ExampleAccount_Under() {
	fmt.Println(Account("Assets:Cash").Under("Assets"))
	fmt.Println(Account("Assets:Cash:Wallet").Under("Assets"))
	fmt.Println(Account("Assets:Cash").Under(""))
	// Output:
	// true
	// true
	// true
}
