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

package month

import (
	"reflect"
	"testing"

	"cloud.google.com/go/civil"
)

func TestNext(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    civil.Date
		want civil.Date
	}{
		{civil.Date{2009, 1, 15}, civil.Date{2009, 2, 1}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.d.String(), func(t *testing.T) {
			t.Parallel()
			got := Next(c.d)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Next() = %#v; want %#v", got, c.want)
			}
		})
	}
}

func TestLastDay(t *testing.T) {
	t.Parallel()
	cases := []struct {
		d    civil.Date
		want civil.Date
	}{
		{civil.Date{2009, 2, 15}, civil.Date{2009, 2, 28}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.d.String(), func(t *testing.T) {
			t.Parallel()
			got := LastDay(c.d)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("LastDay() = %#v; want %#v", got, c.want)
			}
		})
	}
}
