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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper"
)

func TestParse(t *testing.T) {
	t.Parallel()
	r := strings.NewReader(`Equity:Me Assets:Cash 100 USD
Liabilities:CreditCard Expenses:Food 10 USD
Assets:Cash Liabilities:CreditCard 10 USD
Revenues:Income Assets:Cash 20 USD
`)
	got, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}
	want := []Transaction{
		{
			From:     Account{Type: Equity, Name: "Me"},
			To:       Account{Type: Assets, Name: "Cash"},
			Quantity: keeper.NewQuantity(100, 0, "USD"),
		},
		{
			From:     Account{Type: Liabilities, Name: "CreditCard"},
			To:       Account{Type: Expenses, Name: "Food"},
			Quantity: keeper.NewQuantity(10, 0, "USD"),
		},
		{
			From:     Account{Type: Assets, Name: "Cash"},
			To:       Account{Type: Liabilities, Name: "CreditCard"},
			Quantity: keeper.NewQuantity(10, 0, "USD"),
		},
		{
			From:     Account{Type: Revenues, Name: "Income"},
			To:       Account{Type: Assets, Name: "Cash"},
			Quantity: keeper.NewQuantity(20, 0, "USD"),
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
