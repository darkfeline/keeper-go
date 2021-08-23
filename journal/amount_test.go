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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func ExampleBalance_Add() {
	var b Balance
	b.Add(amnt(500, Unit{Symbol: "USD", Scale: 100}))
	fmt.Println(&b)
	// Output:
	// 5.00 USD
}

func TestBalance_Neg(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	var b Balance
	b.Add(amnt(12345, u))
	b.Neg()
	got := b.Amount(u)
	want := amnt(-12345, u)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("amount mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Equal(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	var a Balance
	a.Add(amnt(123, u))
	var b Balance
	b.Add(amnt(124, u))
	if a.Equal(&b) {
		t.Errorf("a.Equal(b) returned true")
	}
}

func TestBalance_Empty_ignore_zero(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	var b Balance
	b.Add(amnt(1, u))
	b.Add(amnt(-1, u))
	if !b.Empty() {
		t.Errorf("Expected b.Empty()")
	}
}

func TestBalance_String(t *testing.T) {
	t.Parallel()
	var b Balance
	b.Add(amnt(-123, Unit{Symbol: "AAA", Scale: 1000}))
	b.Add(amnt(-321, Unit{Symbol: "BBB", Scale: 1000}))
	b.Add(amnt(0, Unit{Symbol: "CCC", Scale: 1000}))
	got := b.String()
	want := "-0.123 AAA, -0.321 BBB"
	if got != want {
		t.Errorf("Got %#v, want %#v", got, want)
	}
}

func TestBalance_Amounts(t *testing.T) {
	t.Parallel()
	var b Balance
	b.Add(amnt(-123, Unit{Symbol: "AAA", Scale: 1000}))
	b.Add(amnt(-321, Unit{Symbol: "BBB", Scale: 1000}))
	b.Add(amnt(0, Unit{Symbol: "CCC", Scale: 1000}))
	got := b.Amounts()
	want := []*Amount{
		amnt(-123, Unit{Symbol: "AAA", Scale: 1000}),
		amnt(-321, Unit{Symbol: "BBB", Scale: 1000}),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Amounts() mismatch (-want +got):\n%s", diff)
	}
}

func amnt(n int64, u Unit) *Amount {
	a := &Amount{Unit: u}
	a.Number.SetInt64(n)
	return a
}
