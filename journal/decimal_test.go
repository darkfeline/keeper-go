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
)

func TestParseDecimal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    string
		want decimal
	}{
		{"123", decimal{123, 1}},
		{"123.", decimal{123, 1}},
		{"123.45", decimal{12345, 100}},
		{"-123.45", decimal{-12345, 100}},
		{"2,123.45", decimal{212345, 100}},
		{"0.02", decimal{2, 100}},
		{"-0.02", decimal{-2, 100}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.s, func(t *testing.T) {
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
