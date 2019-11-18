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

package cmd

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/book"
)

func TestTxStarting(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		ts   []book.Transaction
		d    civil.Date
		want []book.Transaction
	}{
		{
			desc: "happy",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 6},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
		{
			desc: "empty",
			ts:   []book.Transaction{},
			d:    civil.Date{2001, 5, 6},
			want: []book.Transaction{},
		},
		{
			desc: "beginning",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 5},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
		{
			desc: "past beginning",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 4},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
		{
			desc: "end",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 7},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
		{
			desc: "past end",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d:    civil.Date{2001, 5, 8},
			want: []book.Transaction{},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := txStarting(c.ts, c.d)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("tx mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTxEnding(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		ts   []book.Transaction
		d    civil.Date
		want []book.Transaction
	}{
		{
			desc: "happy",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 6},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
			},
		},
		{
			desc: "empty",
			ts:   []book.Transaction{},
			d:    civil.Date{2001, 5, 6},
			want: []book.Transaction{},
		},
		{
			desc: "beginning",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 5},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
			},
		},
		{
			desc: "past beginning",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d:    civil.Date{2001, 5, 4},
			want: []book.Transaction{},
		},
		{
			desc: "end",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 7},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
		{
			desc: "past end",
			ts: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
			d: civil.Date{2001, 5, 8},
			want: []book.Transaction{
				{Date: civil.Date{2001, 5, 5}, Description: "foo"},
				{Date: civil.Date{2001, 5, 6}, Description: "bar"},
				{Date: civil.Date{2001, 5, 7}, Description: "baz"},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := txEnding(c.ts, c.d)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("tx mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
