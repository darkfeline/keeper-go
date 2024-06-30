// Copyright (C) 2024  Allen Li
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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/piquette/finance-go"
	"go.felesatra.moe/keeper/journal"
)

func TestConvertAmount(t *testing.T) {
	t.Parallel()
	q := &finance.Quote{
		RegularMarketPrice: 123.45,
	}
	a := &journal.Amount{Unit: journal.Unit{
		Symbol: "MOE",
		Scale:  1000,
	}}
	a.Number.SetInt64(54321)
	got := convertAmount(a, q)
	want := &journal.Amount{Unit: journal.Unit{
		Symbol: "USD",
		Scale:  100,
	}}
	want.Number.SetInt64(670593)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("amount mismatch (-want +got):\n%s", diff)
	}
}
