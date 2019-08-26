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

package stage1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper"
)

func TestMakeBalance(t *testing.T) {
	t.Parallel()
	ts := []Transaction{
		{
			From:   Account{Type: Equity, Name: "Me"},
			To:     Account{Type: Assets, Name: "Cash"},
			Amount: keeper.Fixed{Value: 100},
			Unit:   "USD",
		},
		{
			From:   Account{Type: Liabilities, Name: "CreditCard"},
			To:     Account{Type: Expenses, Name: "Food"},
			Amount: keeper.Fixed{Value: 10},
			Unit:   "USD",
		},
		{
			From:   Account{Type: Assets, Name: "Cash"},
			To:     Account{Type: Liabilities, Name: "CreditCard"},
			Amount: keeper.Fixed{Value: 10},
			Unit:   "USD",
		},
		{
			From:   Account{Type: Revenues, Name: "Income"},
			To:     Account{Type: Assets, Name: "Cash"},
			Amount: keeper.Fixed{Value: 20},
			Unit:   "USD",
		},
	}
	got := MakeBalance(ts)
	want := Balance{
		Assets: TypeBalance{
			"Cash": {keeper.NewQuantity(110, 0, "USD")},
		},
		Liabilities: TypeBalance{
			"CreditCard": {keeper.NewQuantity(0, 0, "USD")},
		},
		Equity: TypeBalance{
			"Me": {keeper.NewQuantity(-100, 0, "USD")},
		},
		Revenues: TypeBalance{
			"Income": {keeper.NewQuantity(-20, 0, "USD")},
		},
		Expenses: TypeBalance{
			"Food": {keeper.NewQuantity(10, 0, "USD")},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("foo() mismatch (-want +got):\n%s", diff)
	}
}
