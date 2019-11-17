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

package colfmt

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func ExampleFormat() {
	type amount struct {
		number string `colfmt:"right"`
		unit   string
	}
	v := []amount{
		{"123.45", "USD"},
		{"-1.23", "USD"},
		{"18.435", "HOOG"},
	}
	Format(os.Stdout, v)
	// Output:
	// 123.45 USD
	//  -1.23 USD
	// 18.435 HOOG
}

func TestGetSliceColspecs(t *testing.T) {
	t.Parallel()
	type amount struct {
		number string `colfmt:"right"`
		unit   string
	}
	v := []amount{
		{"123.45", "USD"},
		{"-1.23", "USD"},
		{"18.435", "HOOG"},
	}
	got := getSliceColspecs(v)
	want := []colspec{
		{Width: 6, Align: alignRight},
		{Width: 0},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("colspec mismatch (-want +got):\n%s", diff)
	}
}

func TestFormatString(t *testing.T) {
	t.Parallel()
	c := []colspec{
		{Width: 6, Align: alignRight},
		{Width: 0},
	}
	got := formatString(c)
	want := "%6s %s\n"
	if got != want {
		t.Errorf("formatString() = %#v; want %#v", got, want)
	}
}

func BenchmarkFprintf_5_1M(b *testing.B) {
	v := setupStruct5(1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range v {
			fmt.Fprintf(ioutil.Discard, "%s\n", v)
		}
	}
}

func BenchmarkFormat_5_1M(b *testing.B) {
	v := setupStruct5(1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Format(ioutil.Discard, v)
	}
}

type struct5 struct {
	a string
	b string
	c string
	d string
	e string
}

func setupStruct5(n int) []struct5 {
	var v []struct5
	for i := 0; i < n; i++ {
		v = append(v, struct5{"ayanami", "ibuki", "laffey", "gascogne", "cleveland"})
	}
	return v
}