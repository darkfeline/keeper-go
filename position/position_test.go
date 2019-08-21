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

package position

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/fixed"
)

func TestPosition_Add(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		ps   []Position
		f    fixed.Fixed
		u    Unit
		want []Position
	}{
		{"empty", nil,
			fixed.New(123, 1), "USD",
			[]Position{{Amount: fixed.New(123, 1), Unit: "USD"}}},
		{"existing", []Position{{Amount: fixed.New(123, 1), Unit: "USD"}},
			fixed.New(9, 1), "USD",
			[]Position{{Amount: fixed.New(132, 1), Unit: "USD"}}},
		{"different currency", []Position{{Amount: fixed.New(123, 1), Unit: "JPY"}},
			fixed.New(9, 1), "USD",
			[]Position{{Amount: fixed.New(123, 1), Unit: "JPY"},
				{Amount: fixed.New(9, 1), Unit: "USD"}}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := Add(c.ps, c.f, c.u)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("Add() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
