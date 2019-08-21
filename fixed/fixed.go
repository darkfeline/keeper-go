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

package fixed

import (
	"fmt"
	"math"
	"strconv"

	"golang.org/x/xerrors"
)

type Fixed struct {
	Value int64
	Point uint8
}

func New(value int64, point uint8) Fixed {
	return Fixed{
		Value: value,
		Point: point,
	}
}

func Parse(s string) (Fixed, error) {
	p := len(s)
	for i, b := range s {
		if b == '.' {
			p = i
			break
		}
	}
	x, err := strconv.ParseInt(s[:p], 10, 64)
	if err != nil {
		return Fixed{}, xerrors.Errorf("parse decimal %#v: %s", s, err)
	}
	var y int64
	if p+1 < len(s) {
		y, err = strconv.ParseInt(s[p+1:], 10, 64)
		if err != nil {
			return Fixed{}, xerrors.Errorf("parse decimal %#v: %s", s, err)
		}
	}
	p = len(s) - p
	if p > 0 {
		p--
	}
	return Fixed{
		Value: x*int64(math.Pow10(p)) + y,
		Point: uint8(p),
	}, nil
}

func (f Fixed) Add(f2 Fixed) Fixed {
	if f2.Point < f.Point {
		f, f2 = f2, f
	}
	f = f.RaisePoint(f2.Point - f.Point)
	return Fixed{
		Value: f.Value + f2.Value,
		Point: f.Point,
	}
}

func (f Fixed) Neg() Fixed {
	f.Value = -f.Value
	return f
}

func (f Fixed) RaisePoint(n uint8) Fixed {
	return Fixed{
		Value: f.Value * int64(math.Pow10(int(n))),
		Point: f.Point + n,
	}
}

func (f Fixed) String() string {
	if f.Point == 0 {
		return fmt.Sprintf("%d.", f.Value)
	}
	scale := int64(math.Pow10(int(f.Point)))
	return fmt.Sprintf("%d.%d", f.Value/scale, f.Value%scale)
}
