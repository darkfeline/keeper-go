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
	"sort"

	"cloud.google.com/go/civil"
)

// entriesStarting returns the entries with dates on or after the given date.
// The returned slice is sliced from the argument slice.
func entriesStarting(e []Entry, d civil.Date) []Entry {
	i := sort.Search(len(e), func(i int) bool {
		return !e[i].Date().Before(d)
	})
	return e[i:]
}

// entriesEnding returns the entries with dates before or on the given date.
// The returned slice is sliced from the argument slice.
func entriesEnding(e []Entry, d civil.Date) []Entry {
	i := sort.Search(len(e), func(i int) bool {
		return !e[len(e)-1-i].Date().After(d)
	})
	return e[:len(e)-i]
}
