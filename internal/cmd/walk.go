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

	"go.felesatra.moe/keeper/journal"
)

// walkAccountTree calls the given function for every account in the tree of accounts.
// The input slice is sorted in place.
// If parent accounts are missing, they are also visited as virtual nodes.
func walkAccountTree(a []journal.Account, f func(accountNode) error) error {
	sortAccounts(a)
	var last journal.Account
	for i, cur := range a {
		if err := walkBetweenLast(last, cur, f); err != nil {
			return err
		}
		n := accountNode{Account: cur, Leaf: true}
		if i+1 < len(a) {
			if next := a[i+1]; next.Under(cur) {
				n.Leaf = false
			}
		}
		if err := f(n); err != nil {
			return fmt.Errorf("map account tree: %w", err)
		}
		last = cur
	}
	return nil
}

// walkBetweenLast walks the accounts between the last account and
// the current account as virtual nodes.
func walkBetweenLast(last, cur journal.Account, f func(accountNode) error) error {
	parts := cur.Parts()
	for i := len(commonPrefix(last.Parts(), parts)) + 1; i < len(parts); i++ {
		n := accountNode{
			Account: acctConcat(parts[:i]),
			Virtual: true,
		}
		if err := f(n); err != nil {
			return fmt.Errorf("map account tree: %w", err)
		}
	}
	return nil
}

func commonPrefix(a, b []string) []string {
	var prefix []string
	for i, v := range a {
		if i >= len(b) || v != b[i] {
			return prefix
		}
		prefix = append(prefix, v)
	}
	return prefix
}

// accountNode is passed to functions by walkAccountTree.
type accountNode struct {
	Account journal.Account
	// Virtual is true if this Account is a missing parent account.
	Virtual bool
	// Leaf is true if this Account is a leaf in the walked tree.
	Leaf bool
}
