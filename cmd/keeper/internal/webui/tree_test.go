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
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/journal"
)

func TestMakeAccountTree(t *testing.T) {
	t.Parallel()
	type c = []*accountTree
	cases := []struct {
		desc string
		a    []journal.Account
		want *accountTree
	}{
		{
			desc: "simple",
			a:    []journal.Account{"IJN:Ayanami", "USS:Laffey"},
			want: &accountTree{Virtual: true, Children: c{
				{Account: "IJN", Virtual: true, Children: c{
					{Account: "IJN:Ayanami"},
				}},
				{Account: "USS", Virtual: true, Children: c{
					{Account: "USS:Laffey"},
				}},
			}},
		},
		{
			desc: "deep",
			a: []journal.Account{
				"Expenses:Foo:Bar:Baz",
				"Expenses:Spam:Eggs:Ham",
			},
			want: &accountTree{Virtual: true, Children: c{
				{Account: "Expenses", Virtual: true, Children: c{
					{Account: "Expenses:Foo", Virtual: true, Children: c{
						{Account: "Expenses:Foo:Bar", Virtual: true, Children: c{
							{Account: "Expenses:Foo:Bar:Baz"},
						}},
					}},
					{Account: "Expenses:Spam", Virtual: true, Children: c{
						{Account: "Expenses:Spam:Eggs", Virtual: true, Children: c{
							{Account: "Expenses:Spam:Eggs:Ham"},
						}},
					}},
				}},
			}},
		},
		{
			desc: "with parent",
			a: []journal.Account{
				"Expenses:Foo",
				"Expenses:Foo:Bar:Baz",
			},
			want: &accountTree{Virtual: true, Children: c{
				{Account: "Expenses", Virtual: true, Children: c{
					{Account: "Expenses:Foo", Children: c{
						{Account: "Expenses:Foo:Bar", Virtual: true, Children: c{
							{Account: "Expenses:Foo:Bar:Baz"}}},
					}},
				}},
			}},
		},
		{
			desc: "cousins",
			a: []journal.Account{
				"Assets:2019:Foo",
				"Assets:2020:Foo",
			},
			want: &accountTree{Virtual: true, Children: c{
				{Account: "Assets", Virtual: true, Children: c{
					{Account: "Assets:2019", Virtual: true, Children: c{
						{Account: "Assets:2019:Foo"},
					}},
					{Account: "Assets:2020", Virtual: true, Children: c{
						{Account: "Assets:2020:Foo"},
					}},
				}},
			}},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := makeAccountTree(c.a)
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
