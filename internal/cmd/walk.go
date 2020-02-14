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

package cmd

import (
	"fmt"

	"go.felesatra.moe/keeper/book"
)

// walkAccountTree calls the given function for every account in the tree of accounts.
// The input slice is sorted in place.
// If parent accounts are missing, they are also visited as virtual nodes.
func walkAccountTree(a []book.Account, f func(accountNode) error) error {
	sortAccounts(a)
	var last []string
	for _, a := range a {
		parts := acctParts(a)
		common := commonPrefix(last, parts)
		vlen := len(parts)
		for i := len(common) + 1; i < vlen; i++ {
			n := accountNode{
				Account: acctConcat(parts[:i]),
				Virtual: true,
			}
			if err := f(n); err != nil {
				return fmt.Errorf("map account tree: %w", err)
			}
		}
		n := accountNode{
			Account: a,
		}
		if err := f(n); err != nil {
			return fmt.Errorf("map account tree: %w", err)
		}
		last = parts
	}
	return nil
}

func commonPrefix(a, b []string) []string {
	var prefix []string
	for i, v := range a {
		if i >= len(b) {
			return prefix
		}
		if v == b[i] {
			prefix = append(prefix, v)
		}
	}
	return prefix
}

// accountNode is passed to functions by walkAccountTree.
type accountNode struct {
	Account book.Account
	// Virtual is true if this Account is a missing parent account.
	Virtual bool
}
