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
	"sort"
	"strings"
)

// Account is a bookkeeping account.
// Accounts are colon separated paths.
type Account string

// Concat concatenates strings into an Account.
func Concat(a []string, b ...string) Account {
	parts := make([]string, len(a)+len(b))
	copy(parts, a)
	copy(parts[len(a):], b)
	return Account(strings.Join(parts, ":"))
}

// Parts returns the string parts of the Account.
func (a Account) Parts() []string {
	return strings.Split(string(a), ":")
}

// Leaf returns the leaf part of the Account (after the last colon).
func (a Account) Leaf() string {
	p := a.Parts()
	return p[len(p)-1]
}

// Under returns true if the Account is a child account of the argument.
func (a Account) Under(parent Account) bool {
	return strings.HasPrefix(string(a), string(parent)+":")
}

// WalkAccountTree calls the function for every account in the tree of accounts.
// If parent accounts are missing, they are also visited as virtual nodes.
func WalkAccountTree(as []Account, f func(AccountNode) error) error {
	new := make([]Account, len(as))
	copy(new, as)
	sort.Slice(new, func(i, j int) bool { return new[i] < new[j] })
	var last []string
	for _, a := range new {
		parts := a.Parts()
		common := commonPrefix(last, parts)
		vlen := len(parts)
		for i := len(common) + 1; i < vlen; i++ {
			n := AccountNode{
				Account: Concat(parts[:i]),
				Virtual: true,
			}
			if err := f(n); err != nil {
				return fmt.Errorf("map account tree: %w", err)
			}
		}
		n := AccountNode{
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

// AccountNode is passed to functions by WalkAccountTree.
type AccountNode struct {
	Account Account
	// Virtual is true if this Account is a missing parent account.
	Virtual bool
}
