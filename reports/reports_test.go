// Copyright (C) 2022  Allen Li
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

package reports

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/journal"
)

func TestNewAccountLedger_multi_line_entry(t *testing.T) {
	t.Parallel()
	u := journal.Unit{Symbol: "USD", Scale: 100}
	u2 := journal.Unit{Symbol: "BTC", Scale: 100}
	j := &journal.Journal{
		Entries: []journal.Entry{
			&journal.Transaction{
				EntryDate:   civil.Date{2001, 02, 03},
				Description: "test",
				Splits: []journal.Split{
					{Account: "Foo", Amount: amount(123, u)},
					{Account: "Foo", Amount: amount(-4, u2)},
					{Account: "Bar", Amount: amount(-123, u)},
					{Account: "Bar", Amount: amount(4, u2)},
				},
			},
		},
	}
	got := NewAccountLedger(j, "Foo")
	want := []LedgerRow{
		{
			Date:        civil.Date{2001, 02, 03},
			Description: "test",
			Ref:         "-",
			Pair:        Pair[*journal.Amount]{Debit: amount(123, u)},
			Balance:     new(balFac).add(u, 123).bal(),
		},
		{
			Date:        civil.Date{2001, 02, 03},
			Description: "test",
			Ref:         "-",
			Pair:        Pair[*journal.Amount]{Credit: amount(-4, u2)},
			Balance:     new(balFac).add(u, 123).add(u2, -4).bal(),
		},
	}
	if diff := cmpdiff(want, got.Rows); diff != "" {
		t.Errorf("ledger mismatch (-want +got):\n%s", diff)
	}
}

func amount(n int64, u journal.Unit) *journal.Amount {
	a := journal.Amount{
		Unit: u,
	}
	a.Number.SetInt64(n)
	return &a
}

func cmpdiff(x, y interface{}) string {
	return cmp.Diff(x, y, cmpopts...)
}

var cmpopts = []cmp.Option{
	cmp.Comparer(func(x, y journal.Balance) bool {
		return x.Equal(&y)
	}),
}

// Balance factory
type balFac struct {
	b journal.Balance
}

func (f *balFac) bal() journal.Balance {
	return f.b
}

func (f *balFac) pbal() *journal.Balance {
	return &f.b
}

func (f *balFac) add(u journal.Unit, n int64) *balFac {
	a := &journal.Amount{Unit: u}
	a.Number.SetInt64(n)
	f.b.Add(a)
	return f
}
