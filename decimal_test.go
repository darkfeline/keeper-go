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

func TestParseDecimal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		s    string
		want Decimal
	}{
		{"no dot", "123", NewDecimal(123, 0)},
		{"dot at end", "123.", NewDecimal(123, 0)},
		{"dot", "123.45", NewDecimal(12345, 2)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDecimal(c.s)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("ParseDecimal(%#v) = %#v, want %#v", c.s, got, c.want)
			}
		})
	}
}

func TestDecimal_Add(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		x    Decimal
		y    Decimal
		want Decimal
	}{
		{"same", NewDecimal(123, 1), NewDecimal(234, 1), NewDecimal(357, 1)},
		{"first higher point", NewDecimal(123, 2), NewDecimal(234, 1), NewDecimal(2463, 2)},
		{"second higher point", NewDecimal(234, 1), NewDecimal(123, 2), NewDecimal(2463, 2)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
		})
	}
}

func TestDecimal_RaiseExp(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		d    Decimal
		n    uint8
		want Decimal
	}{
		{"zero", NewDecimal(1234, 2), 0, NewDecimal(1234, 2)},
		{"nonzero", NewDecimal(1234, 2), 2, NewDecimal(123400, 4)},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := c.d.RaiseExp(c.n)
			if got != c.want {
				t.Errorf("Got %#v; want %#v", got, c.want)
			}
		})
	}
}

func TestDecimal_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		d    Decimal
		want string
	}{
		{"zero point", NewDecimal(1234, 0), "1234."},
		{"nonzero point", NewDecimal(12345, 3), "12.345"},
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
