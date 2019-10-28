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

package raw

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// Decimal is a floating point number implemented using a scale
// factor.
type Decimal struct {
	Number int64
	// Scale indicates the minimum fractional unit amount,
	// e.g. 100 means 0.01 is the smallest amount.
	// Should be a multiple of 10.
	Scale int64
}

func parseDecimal(s string) (Decimal, error) {
	if len(s) == 0 {
		return Decimal{}, errors.New("parse decimal: empty string")
	}
	p := len(s)
	for i, b := range s {
		if b == '.' {
			p = i
			break
		}
	}
	x, err := strconv.ParseInt(s[:p], 10, 64)
	if err != nil {
		return Decimal{}, fmt.Errorf("parse decimal %#v: %s", s, err)
	}
	var y int64
	if p+1 < len(s) {
		y, err = strconv.ParseInt(s[p+1:], 10, 64)
		if err != nil {
			return Decimal{}, fmt.Errorf("parse decimal %#v: %s", s, err)
		}
	}
	p = len(s) - p
	if p > 0 {
		p--
	}
	if x < 0 {
		y = -y
	}
	scale := int64(math.Pow10(p))
	return Decimal{
		Number: x*scale + y,
		Scale:  scale,
	}, nil
}

func (d Decimal) String() string {
	if d.Scale <= 1 {
		return fmt.Sprintf("%d", d.Number)
	}
	f := d.Number % d.Scale
	if f < 0 {
		f = -f
	}
	return fmt.Sprintf("%d.%d", d.Number/d.Scale, f)
}
