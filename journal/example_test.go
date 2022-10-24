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

package journal_test

import (
	"fmt"
	"sort"

	"go.felesatra.moe/keeper/journal"
)

func ExampleCompile() {
	input := []byte(`unit USD 100
tx 2020-01-01 "Initial balance"
Assets:Cash 100 USD
Equity:Capital
end
`)
	j, err := journal.Compile(&journal.CompileArgs{
		Inputs: []journal.CompileInput{journal.Bytes("input", input)},
	})
	if err != nil {
		panic(err)
	}
	var as []journal.Account
	for a := range j.Accounts {
		as = append(as, a)
	}
	sort.Slice(as, func(i, j int) bool { return as[i] < as[j] })
	for _, a := range as {
		fmt.Println(a)
	}
	// Output:
	// Assets:Cash
	// Equity:Capital
}
