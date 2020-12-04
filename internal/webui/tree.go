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

package webui

import (
	"strings"

	"go.felesatra.moe/keeper/journal"
)

func journalAccountTree(j *journal.Journal) *accountTree {
	var as []journal.Account
	for _, a := range j.Accounts() {
		if j.Disabled[a] == nil {
			as = append(as, a)
		}
	}
	return makeAccountTree(as)
}

type accountTree struct {
	Account  journal.Account
	Virtual  bool
	Children []*accountTree
}

// accounts must be sorted.
func makeAccountTree(a []journal.Account) *accountTree {
	at := &accountTree{
		Virtual: true,
	}
	makeAccountTree1(a, at, 0)
	return at
}

func makeAccountTree1(a []journal.Account, at *accountTree, i int) int {
	cur := at.Account
	curParts := cur.Parts()
	for i < len(a) {
		next := a[i]
		if !next.Under(cur) {
			// Return to a higher frame to handle.
			return i
		}
		nextParts := next.Parts()
		prefix := commonPrefix(curParts, nextParts)
		nextPart := nextParts[len(curParts)]
		at2 := &accountTree{
			Account: journal.Account(strings.Join(append(cur.Parts(), nextPart), ":")),
			Virtual: len(nextParts) > len(prefix)+1,
		}
		at.Children = append(at.Children, at2)
		if at2.Virtual {
			i = makeAccountTree1(a, at2, i)
		} else {
			i = makeAccountTree1(a, at2, i+1)
		}
	}
	return i
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
