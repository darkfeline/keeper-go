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

package parse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/book"
)

func TestAcctBalance_Add(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	var b acctBalance
	b.Add(book.Amount{Number: 12345, UnitType: u})
	want := acctBalance{{Number: 12345, UnitType: u}}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestAcctBalance_Add_existing_balance(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	b := acctBalance{{Number: 10000, UnitType: u}}
	b.Add(book.Amount{Number: 12345, UnitType: u})
	want := acctBalance{{Number: 22345, UnitType: u}}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestAcctBalance_Equal(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	u2 := &book.UnitType{Symbol: "JPY", Scale: 1}
	a := acctBalance{
		{Number: 123, UnitType: u},
		{Number: 3200, UnitType: u2},
	}
	b := acctBalance{
		{Number: 3200, UnitType: u2},
		{Number: 123, UnitType: u},
	}
	if !a.Equal(b) {
		t.Errorf("a.Equal(b) returned false")
	}
}

func TestAcctBalance_Equal_different_length(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	u2 := &book.UnitType{Symbol: "JPY", Scale: 1}
	a := acctBalance{
		{Number: 3200, UnitType: u2},
	}
	b := acctBalance{
		{Number: 123, UnitType: u},
		{Number: 3200, UnitType: u2},
	}
	if a.Equal(b) {
		t.Errorf("a.Equal(b) returned true")
	}
	if b.Equal(a) {
		t.Errorf("b.Equal(a) returned true")
	}
}

func TestAcctBalance_Equal_different_amount(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	u2 := &book.UnitType{Symbol: "JPY", Scale: 1}
	a := acctBalance{
		{Number: 3200, UnitType: u2},
		{Number: 123, UnitType: u},
	}
	b := acctBalance{
		{Number: 123, UnitType: u},
		{Number: 200, UnitType: u2},
	}
	if a.Equal(b) {
		t.Errorf("a.Equal(b) returned true")
	}
}

func TestAcctBalance_Equal_ignore_zero(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	u2 := &book.UnitType{Symbol: "JPY", Scale: 1}
	a := acctBalance{
		{Number: 3200, UnitType: u2},
		{Number: 0, UnitType: u},
	}
	b := acctBalance{
		{Number: 3200, UnitType: u2},
	}
	if !a.Equal(b) {
		t.Errorf("a.Equal(b) returned false")
	}
	if !b.Equal(a) {
		t.Errorf("b.Equal(a) returned false")
	}
}

func TestAcctBalance_Add_zeroed_accounts(t *testing.T) {
	t.Parallel()
	u := &book.UnitType{Symbol: "USD", Scale: 100}
	u2 := &book.UnitType{Symbol: "JPY", Scale: 1}
	b := acctBalance{
		{Number: 3200, UnitType: u2},
		{Number: 123, UnitType: u},
	}
	b.Add(book.Amount{Number: -123, UnitType: u})
	want := acctBalance{
		{Number: 3200, UnitType: u2},
	}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}
