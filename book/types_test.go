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

package book

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func ExampleAccount_Parts() {
	fmt.Printf("%#v\n", Account("Assets:Cash").Parts())
	fmt.Printf("%#v\n", Account("Assets").Parts())
	// Output:
	// []string{"Assets", "Cash"}
	// []string{"Assets"}
}

func ExampleAccount_Level() {
	fmt.Println(Account("Assets:Cash").Level())
	fmt.Println(Account("Assets").Level())
	// Output:
	// 2
	// 1
}

func ExampleAccount_Parent() {
	fmt.Printf("%#v\n", Account("Assets:Cash").Parent())
	fmt.Printf("%#v\n", Account("Assets").Parent())
	// Output:
	// "Assets"
	// ""
}

func ExampleAccount_Leaf() {
	fmt.Printf("%#v\n", Account("Assets:Cash").Leaf())
	fmt.Printf("%#v\n", Account("Assets").Leaf())
	// Output:
	// "Cash"
	// "Assets"
}

func ExampleAccount_Under() {
	fmt.Println(Account("Assets:Cash").Under("Assets"))
	fmt.Println(Account("Assets:Cash:Wallet").Under("Assets"))
	// Output:
	// true
	// true
}

func TestBalance_Add_empty_balance(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	var b Balance
	b = b.Add(Amount{Number: 12345, Unit: u})
	want := Balance{{Number: 12345, Unit: u}}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Add_existing_balance(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	b := Balance{{Number: 10000, Unit: u}}
	b = b.Add(Amount{Number: 12345, Unit: u})
	want := Balance{{Number: 22345, Unit: u}}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Add_zeroed_accounts(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	b := Balance{
		{Number: 3200, Unit: u2},
		{Number: 123, Unit: u},
	}
	b = b.Add(Amount{Number: -123, Unit: u})
	want := Balance{
		{Number: 3200, Unit: u2},
	}
	if diff := cmp.Diff(want, b); diff != "" {
		t.Errorf("balance mismatch (-want +got):\n%s", diff)
	}
}

func TestBalance_Equal_ignores_order(t *testing.T) {
	t.Parallel()
	u := Unit{Symbol: "USD", Scale: 100}
	u2 := Unit{Symbol: "JPY", Scale: 1}
	a := Balance{
		{Number: 123, Unit: u},
		{Number: 3200, Unit: u2},
	}
	b := Balance{
		{Number: 3200, Unit: u2},
		{Number: 123, Unit: u},
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
		{Number: 3200, Unit: u2},
	}
	b := Balance{
		{Number: 123, Unit: u},
		{Number: 3200, Unit: u2},
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
		{Number: 3200, Unit: u2},
		{Number: 123, Unit: u},
	}
	b := Balance{
		{Number: 123, Unit: u},
		{Number: 200, Unit: u2},
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
		{Number: 3200, Unit: u2},
		{Number: 0, Unit: u},
	}
	b := Balance{
		{Number: 3200, Unit: u2},
	}
	if !a.Equal(b) {
		t.Errorf("a.Equal(b) returned false")
	}
	if !b.Equal(a) {
		t.Errorf("b.Equal(a) returned false")
	}
}

func TestBalance_CleanCopy(t *testing.T) {
	t.Parallel()
	b := Balance{
		Amount{Number: 0, Unit: Unit{Symbol: "VTRA", Scale: 1000}},
		Amount{Number: 3456, Unit: Unit{Symbol: "VTRB", Scale: 1000}},
	}
	got := b.CleanCopy()
	want := Balance{
		Amount{Number: 3456, Unit: Unit{Symbol: "VTRB", Scale: 1000}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("b.CleanCopy() = %#v; want %#v", got, want)
	}
}
