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

import "testing"

func TestParseFixed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		s    string
		want Fixed
	}{
		{"no dot", "123", NewFixed(123, 0)},
		{"dot at end", "123.", NewFixed(123, 0)},
		{"dot", "123.45", NewFixed(12345, 2)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got, err := ParseFixed(c.s)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("ParseFixed(%#v) = %#v, want %#v", c.s, got, c.want)
			}
		})
	}
}

func TestFixed_Add(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		x    Fixed
		y    Fixed
		want Fixed
	}{
		{"same", NewFixed(123, 1), NewFixed(234, 1), NewFixed(357, 1)},
		{"first higher point", NewFixed(123, 2), NewFixed(234, 1), NewFixed(2463, 2)},
		{"second higher point", NewFixed(234, 1), NewFixed(123, 2), NewFixed(2463, 2)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
		})
	}
}

func TestFixed_RaisePoint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		d    Fixed
		n    uint8
		want Fixed
	}{
		{"zero", NewFixed(1234, 2), 0, NewFixed(1234, 2)},
		{"nonzero", NewFixed(1234, 2), 2, NewFixed(123400, 4)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := c.d.RaisePoint(c.n)
			if got != c.want {
				t.Errorf("Got %#v; want %#v", got, c.want)
			}
		})
	}
}

func TestFixed_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		d    Fixed
		want string
	}{
		{"zero point", NewFixed(1234, 0), "1234."},
		{"nonzero point", NewFixed(12345, 3), "12.345"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := c.d.String()
			if got != c.want {
				t.Errorf("Got %#v, want %#v", got, c.want)
			}
		})
	}
}
