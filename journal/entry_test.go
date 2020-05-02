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

package journal

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
)

func TestSortEntries(t *testing.T) {
	t.Parallel()
	got := []Entry{
		&BalanceAssert{EntryDate: civil.Date{2000, 1, 5}},
		&DisableAccount{EntryDate: civil.Date{2000, 1, 5}},
		&Transaction{EntryDate: civil.Date{2000, 1, 5}},
	}
	sortEntries(got)
	want := []Entry{
		&Transaction{EntryDate: civil.Date{2000, 1, 5}},
		&BalanceAssert{EntryDate: civil.Date{2000, 1, 5}},
		&DisableAccount{EntryDate: civil.Date{2000, 1, 5}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}
