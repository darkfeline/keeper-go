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

package keeper

import (
	"sort"
	"strings"

	"golang.org/x/xerrors"
)

type Account string

func Concat(a []string, b ...string) Account {
	parts := make([]string, len(a)+len(b))
	copy(parts, a)
	copy(parts[len(a):], b)
	return Account(strings.Join(parts, ":"))
}

func (a Account) Parts() []string {
	return strings.Split(string(a), ":")
}

func (a Account) Leaf() string {
	p := a.Parts()
	return p[len(p)-1]
}

func (a Account) Under(parent Account) bool {
	return strings.HasPrefix(string(a), string(parent)+":")
}

func MapAccountTree(as []Account, f func(AccountNode) error) error {
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
				return xerrors.Errorf("map account tree: %w", err)
			}
		}
		n := AccountNode{
			Account: a,
		}
		if err := f(n); err != nil {
			return xerrors.Errorf("map account tree: %w", err)
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

type AccountNode struct {
	Account Account
	Virtual bool
}
