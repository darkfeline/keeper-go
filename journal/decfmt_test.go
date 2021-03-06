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
	"fmt"
	"testing"
)

func TestDecFormat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		n     int64
		scale int64
		want  string
	}{
		{1234, 1, "1,234"},
		{12345, 100, "123.45"},
		{-12345, 100, "-123.45"},
		{10000, 100, "100.00"},
		{12345678, 1, "12,345,678"},
		{4, 100, "0.04"},
		{0, 1, "0"},
		{0, 100, "0.00"},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("n=%d scale=%d", c.n, c.scale), func(t *testing.T) {
			t.Parallel()
			got := decFormat(c.n, c.scale)
			if got != c.want {
				t.Errorf("Format(%v, %v) = %#v; want %#v", c.n, c.scale, got, c.want)
			}
		})
	}
}
