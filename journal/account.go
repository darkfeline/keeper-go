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

import "strings"

// Account is a bookkeeping account.
// Accounts are colon separated paths, like "Income:Salary".
type Account string

// Parts returns the parts of the account between the colons.
func (a Account) Parts() []string {
	if a == "" {
		return nil
	}
	return strings.Split(string(a), ":")
}

// Parent returns the parent account.
func (a Account) Parent() Account {
	p := a.Parts()
	if len(p) == 0 {
		return ""
	}
	return Account(strings.Join(p[:len(p)-1], ":"))
}

// Leaf returns the leaf part of the Account (after the last colon).
func (a Account) Leaf() string {
	if a == "" {
		return ""
	}
	p := a.Parts()
	return p[len(p)-1]
}

// Under returns true if the Account is a child account of the argument.
func (a Account) Under(parent Account) bool {
	if parent == "" {
		return true
	}
	return strings.HasPrefix(string(a), string(parent)+":")
}

// An AccountInfo holds account information.
type AccountInfo struct {
	// If the account is disabled, points to the entry that
	// disabled the account.  Otherwise this is nil.
	Disabled *DisableAccount
	Metadata map[string]string
}
