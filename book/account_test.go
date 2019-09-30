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
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestConcat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		a, b []string
		want Account
	}{
		{"happy", []string{"ayanami"}, []string{"laffey"}, "ayanami:laffey"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := Concat(c.a, c.b...)
			if got != c.want {
				t.Errorf("Concat(%v, %v) = %v; want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestWalkAccountTree(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		a    []Account
		want []AccountNode
	}{
		{
			desc: "simple",
			a:    []Account{"IJN:Ayanami", "USS:Laffey"},
			want: []AccountNode{
				{Account: "IJN", Virtual: true},
				{Account: "IJN:Ayanami"},
				{Account: "USS", Virtual: true},
				{Account: "USS:Laffey"},
			},
		},
		{
			desc: "deep",
			a: []Account{
				"Expenses:Foo:Bar:Baz",
				"Expenses:Spam:Eggs:Ham",
			},
			want: []AccountNode{
				{Account: "Expenses", Virtual: true},
				{Account: "Expenses:Foo", Virtual: true},
				{Account: "Expenses:Foo:Bar", Virtual: true},
				{Account: "Expenses:Foo:Bar:Baz"},
				{Account: "Expenses:Spam", Virtual: true},
				{Account: "Expenses:Spam:Eggs", Virtual: true},
				{Account: "Expenses:Spam:Eggs:Ham"},
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			var got []AccountNode
			f := func(n AccountNode) error {
				got = append(got, n)
				return nil
			}
			if err := WalkAccountTree(c.a, f); err != nil {
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
