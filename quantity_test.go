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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAddUnits(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc string
		ps   []Quantity
		f    Fixed
		u    Unit
		want []Quantity
	}{
		{"empty", nil,
			NewFixed(123, 1), "USD",
			[]Quantity{NewQuantity(123, 1, "USD")}},
		{"existing", []Quantity{NewQuantity(123, 1, "USD")},
			NewFixed(9, 1), "USD",
			[]Quantity{NewQuantity(132, 1, "USD")}},
		{"different currency", []Quantity{NewQuantity(132, 1, "JPY")},
			NewFixed(9, 1), "USD",
			[]Quantity{
				NewQuantity(132, 1, "JPY"),
				NewQuantity(9, 1, "USD"),
			}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got := AddUnits(c.ps, c.f, c.u)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("Add() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
