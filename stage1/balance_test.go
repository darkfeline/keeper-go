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
	"go.felesatra.moe/keeper/fixed"
)

func TestMakeBalance(t *testing.T) {
	t.Parallel()
	ts := []Transaction{
		{
			From:   Account{Type: Equity, Name: "Me"},
			To:     Account{Type: Assets, Name: "Cash"},
			Amount: fixed.Fixed{Value: 100},
			Unit:   "USD",
		},
		{
			From:   Account{Type: Liabilities, Name: "CreditCard"},
			To:     Account{Type: Expenses, Name: "Food"},
			Amount: fixed.Fixed{Value: 10},
			Unit:   "USD",
		},
		{
			From:   Account{Type: Assets, Name: "Cash"},
			To:     Account{Type: Liabilities, Name: "CreditCard"},
			Amount: fixed.Fixed{Value: 10},
			Unit:   "USD",
		},
		{
			From:   Account{Type: Revenues, Name: "Income"},
			To:     Account{Type: Assets, Name: "Cash"},
			Amount: fixed.Fixed{Value: 20},
			Unit:   "USD",
		},
	}
	got := MakeBalance(ts)
	want := Balance{
		Assets: TypeBalance{
			"Cash": {{Amount: fixed.New(110, 0), Unit: "USD"}},
		},
		Liabilities: TypeBalance{
			"CreditCard": {{Amount: fixed.New(0, 0), Unit: "USD"}},
		},
		Equity: TypeBalance{
			"Me": {{Amount: fixed.New(-100, 0), Unit: "USD"}},
		},
		Revenues: TypeBalance{
			"Income": {{Amount: fixed.New(-20, 0), Unit: "USD"}},
		},
		Expenses: TypeBalance{
			"Food": {{Amount: fixed.New(10, 0), Unit: "USD"}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("foo() mismatch (-want +got):\n%s", diff)
	}
}
