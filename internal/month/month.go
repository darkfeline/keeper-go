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

func LastDay(d civil.Date) civil.Date {
	next := d.AddDays(1)
	for next.Month == d.Month {
		d = next
		next = d.AddDays(1)
	}
	return d
}
