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

package journal

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
)

func TestEntriesEnding(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		ts   []Entry
		d    civil.Date
		want []Entry
	}{
		{
			desc: "happy",
			ts: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 6},
			want: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
			},
		},
		{
			desc: "empty",
			ts:   []Entry{},
			d:    civil.Date{2001, 5, 6},
			want: []Entry{},
		},
		{
			desc: "beginning",
			ts: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 5},
			want: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
			},
		},
		{
			desc: "past beginning",
			ts: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d:    civil.Date{2001, 5, 4},
			want: []Entry{},
		},
		{
			desc: "end",
			ts: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 7},
			want: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
		{
			desc: "past end",
			ts: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 8},
			want: []Entry{
				&Transaction{EntryDate: civil.Date{2001, 5, 5}, Description: "foo"},
				&Transaction{EntryDate: civil.Date{2001, 5, 6}, Description: "bar"},
				&Transaction{EntryDate: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := entriesEnding(c.ts, c.d)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("entries mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
