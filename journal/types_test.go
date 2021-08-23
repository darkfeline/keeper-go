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

package journal

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func ExampleBalance_Add() {
	b := make(Balance)
	b.Add(amnt(500, Unit{Symbol: "USD", Scale: 100}))
	fmt.Println(b)
	// Output:
	// 5.00 USD
}

func TestBalance_Add_empty_balance(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	b := make(Balance)
	b.Add(amnt(12345, u))
	want := Balance{u: big.NewInt(12345)}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Add_existing_balance(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	b := Balance{u: big.NewInt(10000)}
	b.Add(amnt(12345, u))
	want := Balance{u: big.NewInt(22345)}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Add_zeroed_accounts(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	b := Balance{
		u2: big.NewInt(3200),
		u:  big.NewInt(123),
	}
	b.Add(amnt(-123, u))
	want := Balance{u2: big.NewInt(3200)}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Neg(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	b := Balance{u: big.NewInt(12345)}
	b.Neg()
	want := Balance{u: big.NewInt(-12345)}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Equal_ignores_order(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	a := Balance{
		u:  big.NewInt(123),
		u2: big.NewInt(3200),
	}
	b := Balance{
		u:  big.NewInt(123),
		u2: big.NewInt(3200),
	}
	if !a.Equal(b) {
		t.Errorf("a.Equal(b) returned false")
	}
}

func TestBalance_Equal_different_length(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	a := Balance{
		u2: big.NewInt(3200),
	}
	b := Balance{
		u:  big.NewInt(123),
		u2: big.NewInt(3200),
	}
	if a.Equal(b) {
		t.Errorf("a.Equal(b) returned true")
	}
	if b.Equal(a) {
		t.Errorf("b.Equal(a) returned true")
	}
}

func TestBalance_Equal_different_amount(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	a := Balance{
		u2: big.NewInt(3200),
		u:  big.NewInt(123),
	}
	b := Balance{
		u:  big.NewInt(123),
		u2: big.NewInt(200),
	}
	if a.Equal(b) {
		t.Errorf("a.Equal(b) returned true")
	}
}

func TestBalance_Equal_ignore_zero(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	a := Balance{
		u2: big.NewInt(3200),
		u:  big.NewInt(0),
	}
	b := Balance{
		u2: big.NewInt(3200),
	}
	if !a.Equal(b) {
		t.Errorf("a.Equal(b) returned false")
	}
	if !b.Equal(a) {
		t.Errorf("b.Equal(a) returned false")
	}
}

func TestBalance_String(t *testing.T) {
	t.Parallel()
	b := Balance{
		Unit{Symbol: "AAA", Scale: 1000}: big.NewInt(-123),
		Unit{Symbol: "BBB", Scale: 1000}: big.NewInt(-321),
		Unit{Symbol: "CCC", Scale: 1000}: big.NewInt(0),
	}
	got := b.String()
	want := "-0.123 AAA, -0.321 BBB"
	if got != want {
		t.Errorf("Got %#v, want %#v", got, want)
	}
}

func TestBalance_Amounts(t *testing.T) {
	t.Parallel()
	b := Balance{
		Unit{Symbol: "AAA", Scale: 1000}: big.NewInt(-123),
		Unit{Symbol: "BBB", Scale: 1000}: big.NewInt(-321),
		Unit{Symbol: "CCC", Scale: 1000}: big.NewInt(0),
	}
	got := b.Amounts()
	want := []*Amount{
		amnt(-123, Unit{Symbol: "AAA", Scale: 1000}),
		amnt(-321, Unit{Symbol: "BBB", Scale: 1000}),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Amounts() mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Copy_nil(t *testing.T) {
	t.Parallel()
	var b Balance
	got := b.Copy()
	if got == nil {
		t.Errorf("Got nil Balance copy")
	}
}

func amnt(n int64, u Unit) *Amount {
	a := &Amount{Unit: u}
	a.Number.SetInt64(n)
	return a
}
