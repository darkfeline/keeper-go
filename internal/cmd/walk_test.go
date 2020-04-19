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

package cmd

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/journal"
)

func TestWalkAccountTree(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		a    []journal.Account
		want []accountNode
	}{
		{
			desc: "simple",
			a:    []journal.Account{"IJN:Ayanami", "USS:Laffey"},
			want: []accountNode{
				{Account: "IJN", Virtual: true},
				{Account: "IJN:Ayanami", Leaf: true},
				{Account: "USS", Virtual: true},
				{Account: "USS:Laffey", Leaf: true},
			},
		},
		{
			desc: "deep",
			a: []journal.Account{
				"Expenses:Foo:Bar:Baz",
				"Expenses:Spam:Eggs:Ham",
			},
			want: []accountNode{
				{Account: "Expenses", Virtual: true},
				{Account: "Expenses:Foo", Virtual: true},
				{Account: "Expenses:Foo:Bar", Virtual: true},
				{Account: "Expenses:Foo:Bar:Baz", Leaf: true},
				{Account: "Expenses:Spam", Virtual: true},
				{Account: "Expenses:Spam:Eggs", Virtual: true},
				{Account: "Expenses:Spam:Eggs:Ham", Leaf: true},
			},
		},
		{
			desc: "with parent",
			a: []journal.Account{
				"Expenses:Foo",
				"Expenses:Foo:Bar:Baz",
			},
			want: []accountNode{
				{Account: "Expenses", Virtual: true},
				{Account: "Expenses:Foo"},
				{Account: "Expenses:Foo:Bar", Virtual: true},
				{Account: "Expenses:Foo:Bar:Baz", Leaf: true},
			},
		},
		{
			desc: "bug 1",
			a: []journal.Account{
				"Assets:2019:Foo",
				"Assets:2020:Foo",
			},
			want: []accountNode{
				{Account: "Assets", Virtual: true},
				{Account: "Assets:2019", Virtual: true},
				{Account: "Assets:2019:Foo", Leaf: true},
				{Account: "Assets:2020", Virtual: true},
				{Account: "Assets:2020:Foo", Leaf: true},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			var got []accountNode
			f := func(n accountNode) error {
				got = append(got, n)
				return nil
			}
			if err := walkAccountTree(c.a, f); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCommonPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		a, b []string
		want []string
	}{
		{"empty a", nil, []string{"foo"}, nil},
		{"empty b", []string{"foo"}, nil, nil},
		{"longer a", []string{"foo", "bar"}, []string{"foo"}, []string{"foo"}},
		{"longer b", []string{"foo"}, []string{"foo", "bar"}, []string{"foo"}},
		{"diverge", []string{"foo", "baz"}, []string{"foo", "bar"}, []string{"foo"}},
		{"same", []string{"foo", "bar"}, []string{"foo", "bar"}, []string{"foo", "bar"}},
		{"bug 1", []string{"Assets", "2019", "Foo"},
			[]string{"Assets", "2020", "Foo"}, []string{"Assets"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := commonPrefix(c.a, c.b)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("commonPrefix(%v, %v) = %v; want %v", c.a, c.b, got, c.want)
			}
		})
	}
}
