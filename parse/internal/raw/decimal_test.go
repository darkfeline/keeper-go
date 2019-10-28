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

package raw

import "testing"

func TestParseDecimal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		s    string
		want Decimal
	}{
		{"no dot", "123", Decimal{123, 1}},
		{"dot at end", "123.", Decimal{123, 1}},
		{"dot", "123.45", Decimal{12345, 100}},
		{"negative", "-123.45", Decimal{-12345, 100}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got, err := parseDecimal(c.s)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("parseDecimal(%#v) = %#v, want %#v", c.s, got, c.want)
			}
		})
	}
}

func TestDeciaml_String(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		d    Decimal
		want string
	}{
		{"unit", Decimal{1234, 1}, "1234"},
		{"fractions", Decimal{12345, 100}, "123.45"},
		{"negative fractions", Decimal{-12345, 100}, "-123.45"},
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
