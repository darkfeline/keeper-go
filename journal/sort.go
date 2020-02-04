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

package book

import (
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/journal"
)

func SortByDate(e []Entry) {
	type pair struct {
		k int64
		v Entry
	}
	ks := make([]pair, len(e))
	for i, e := range e {
		ks[i] = pair{entryKey(e), e}
	}
	sort.Slice(ks, func(i, j int) bool {
		return ks[i].k < ks[j].k
	})
	for i, k := range ks {
		e[i] = k.v
	}
}

// entryKey returns a sort key corresponding to an journal.
func entryKey(e Entry) int64 {
	switch e := e.(type) {
	case journal.Transaction:
		return dateKey(e.Date())
	case journal.BalanceAssert:
		return dateKey(e.Date()) + 1
	default:
		panic(fmt.Sprintf("unknown Entry type %T", e))
	}
}

// dateKey returns a sort key corresponding to a Date.
func dateKey(d civil.Date) int64 {
	return d.In(time.UTC).Unix()
}
