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

// Package month provides utilities for handling months.
package month

import (
	"fmt"
	"time"

	"cloud.google.com/go/civil"
)

func Parse(s string) (civil.Date, error) {
	d, err := civil.ParseDate(s + "-01")
	if err != nil {
		return civil.Date{}, fmt.Errorf("parse month: %s", err)
	}
	return d, nil
}

func Format(d civil.Date) string {
	return fmt.Sprintf("%04d-%02d", d.Year, d.Month)
}

func Now() civil.Date {
	d := civil.DateOf(time.Now())
	return FirstDay(d)
}

func FirstDay(d civil.Date) civil.Date {
	d.Day = 1
	return d
}

func LastDay(d civil.Date) civil.Date {
	return Next(d).AddDays(-1)
}

// Next returns the first day of the next month.
func Next(d civil.Date) civil.Date {
	cur := d.Month
	for d.Month == cur {
		d = d.AddDays(21)
	}
	return FirstDay(d)
}

// Prev returns the first day of the previous month.
func Prev(d civil.Date) civil.Date {
	cur := d.Month
	for d.Month == cur {
		d = d.AddDays(-21)
	}
	return FirstDay(d)
}
