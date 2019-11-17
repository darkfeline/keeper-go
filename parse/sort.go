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

package parse

import (
	"sort"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/parse/raw"
)

func sortEntries(e []interface{}) {
	type keyed struct {
		k int64
		v interface{}
	}
	ks := make([]keyed, len(e))
	for i, e := range e {
		ks[i] = keyed{entryKey(e), e}
	}
	sort.Slice(ks, func(i, j int) bool {
		return ks[i].k < ks[j].k
	})
	for i, k := range ks {
		e[i] = k.v
	}
}

// entryKey returns a sort key corresponding to an entry.
func entryKey(e interface{}) int64 {
	switch e := e.(type) {
	case raw.TransactionEntry:
		return dateKey(e.Date)
	case raw.BalanceEntry:
		return dateKey(e.Date) + 1
	default:
		return 0
	}
}

// dateKey returns a sort key corresponding to a Date.
func dateKey(d civil.Date) int64 {
	return d.In(time.UTC).Unix()
}
