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

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/book"
)

func TestConvertAmount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    decimal
		u    book.UnitType
		want int64
	}{
		{decimal{5, 1000}, book.UnitType{Symbol: "Foo", Scale: 1000}, 5},
		{decimal{5, 10}, book.UnitType{Symbol: "Foo", Scale: 1000}, 500},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%v %v", c.d, c.u), func(t *testing.T) {
			t.Parallel()
			got, err := convertAmount(c.d, c.u)
			if err != nil {
				t.Error(err)
			}
			want := book.Amount{
				Number:   c.want,
				UnitType: c.u,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("amount mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
