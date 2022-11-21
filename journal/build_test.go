// Copyright (C) 2020  Allen Li
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

	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/kpr/token"
)

func TestBuildEntries_simple(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
balance 2001-02-03 Some:account -1.20 USD
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff
end
account Some:account
meta "nilou" "nahida"
end
`
	b, got, err := parseAndBuild(inputBytes{"", []byte(input)})
	if err != nil {
		t.Fatal(err)
	}
	u := Unit{Symbol: "USD", Scale: 100}
	want := []Entry{
		&BalanceAssert{
			EntryDate: civil.Date{2001, 2, 3},
			EntryPos:  token.Position{Offset: 13, Line: 2, Column: 1},
			Account:   "Some:account",
			Declared:  new(balFac).add(u, -120).bal(),
		},
		&Transaction{
			EntryDate:   civil.Date{2001, 2, 3},
			EntryPos:    token.Position{Offset: 55, Line: 3, Column: 1},
			Description: "Buy stuff",
			Splits: []Split{
				split("Some:account", -120, u),
				split("Expenses:Stuff", 120, u),
			},
		},
	}
	if diff := cmpdiff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
	want2 := map[Account]*AccountInfo{
		"Some:account": {
			Metadata: map[string]string{"nilou": "nahida"},
		},
	}
	if diff := cmpdiff(want2, b.accounts); diff != "" {
		t.Errorf("account info mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildEntries_account_merge(t *testing.T) {
	t.Parallel()
	const input = `account Some:account
meta "bocchi" "rock"
meta "nilou" "nahida"
end
account Some:account
meta "mir" "jakuri"
meta "nilou" "kokomi"
end
`
	b, _, err := parseAndBuild(inputBytes{"", []byte(input)})
	if err != nil {
		t.Fatal(err)
	}
	want := map[Account]*AccountInfo{
		"Some:account": {
			Metadata: map[string]string{
				"bocchi": "rock",
				"mir":    "jakuri",
				"nilou":  "kokomi",
			},
		},
	}
	if diff := cmpdiff(want, b.accounts); diff != "" {
		t.Errorf("account info mismatch (-want +got):\n%s", diff)
	}
}

func TestBuildEntries_unbalanced(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
tx 2001-02-03 "Buy stuff"
Some:account -1.2 USD
Expenses:Stuff 1.3 USD
end
`
	_, _, err := parseAndBuild(inputBytes{"", []byte(input)})
	if err == nil {
		t.Errorf("Expected errors")
	}
}

func TestBuildEntries_same_duplicate_unit(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
unit USD 100
`
	_, _, err := parseAndBuild(inputBytes{"", []byte(input)})
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
	}
}

func TestBuildEntries_diff_duplicate_unit(t *testing.T) {
	t.Parallel()
	const input = `unit USD 100
unit USD 10
`
	_, _, err := parseAndBuild(inputBytes{"", []byte(input)})
	if err == nil {
		t.Errorf("Expected error")
	}
}

func TestBuildEntries_disable(t *testing.T) {
	t.Parallel()
	const input = `disable 2001-02-03 Some:account
`
	_, got, err := parseAndBuild(inputBytes{"", []byte(input)})
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{
		&DisableAccount{
			EntryDate: civil.Date{2001, 2, 3},
			EntryPos:  token.Position{Offset: 0, Line: 1, Column: 1},
			Account:   "Some:account",
		},
	}
	if diff := cmpdiff(want, got); diff != "" {
		t.Errorf("entries mismatch (-want +got):\n%s", diff)
	}
}

func TestIsPower10(t *testing.T) {
	t.Parallel()
	cases := []struct {
		n    uint64
		want bool
	}{
		{0, false},
		{11, false},
		{101, false},
		{1, true},
		{10, true},
		{100, true},
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d", c.n), func(t *testing.T) {
			t.Parallel()
			got := isPower10(c.n)
			if got != c.want {
				t.Errorf("isPower10(%d) = %v; want %v", c.n, got, c.want)
			}
		})
	}
}

func parseAndBuild(inputs ...CompileInput) (*builder, []Entry, error) {
	fset := token.NewFileSet()
	e, err := parseEntries(fset, inputs...)
	if err != nil {
		return nil, nil, err
	}
	b := newBuilder(fset)
	e2, err := b.build(e...)
	if err != nil {
		return b, nil, err
	}
	return b, e2, nil
}

func cmpdiff(x, y interface{}) string {
	return cmp.Diff(x, y, cmpopts...)
}

var cmpopts = []cmp.Option{
	cmp.Comparer(func(x, y Balance) bool {
		return x.Equal(&y)
	}),
}

// Balance factory
type balFac struct {
	b Balance
}

func (f *balFac) bal() Balance {
	return f.b
}

func (f *balFac) pbal() *Balance {
	return &f.b
}

func (f *balFac) add(u Unit, n int64) *balFac {
	a := &Amount{Unit: u}
	a.Number.SetInt64(n)
	f.b.Add(a)
	return f
}
