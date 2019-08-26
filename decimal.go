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

package keeper

import (
	"fmt"
	"math"
	"strconv"

	"golang.org/x/xerrors"
)

type Decimal struct {
	Sig int64
	Exp uint8
}

func NewDecimal(sig int64, exp uint8) Decimal {
	return Decimal{
		Sig: sig,
		Exp: exp,
	}
}

func ParseDecimal(s string) (Decimal, error) {
	p := len(s)
	for i, b := range s {
		if b == '.' {
			p = i
			break
		}
	}
	x, err := strconv.ParseInt(s[:p], 10, 64)
	if err != nil {
		return Decimal{}, xerrors.Errorf("parse decimal %#v: %s", s, err)
	}
	var y int64
	if p+1 < len(s) {
		y, err = strconv.ParseInt(s[p+1:], 10, 64)
		if err != nil {
			return Decimal{}, xerrors.Errorf("parse decimal %#v: %s", s, err)
		}
	}
	p = len(s) - p
	if p > 0 {
		p--
	}
	return Decimal{
		Sig: x*int64(math.Pow10(p)) + y,
		Exp: uint8(p),
	}, nil
}

func (d Decimal) Add(d2 Decimal) Decimal {
	if d2.Exp < d.Exp {
		d, d2 = d2, d
	}
	d = d.RaiseExp(d2.Exp - d.Exp)
	return Decimal{
		Sig: d.Sig + d2.Sig,
		Exp: d.Exp,
	}
}

func (d Decimal) Neg() Decimal {
	d.Sig = -d.Sig
	return d
}

func (d Decimal) RaiseExp(n uint8) Decimal {
	return Decimal{
		Sig: d.Sig * int64(math.Pow10(int(n))),
		Exp: d.Exp + n,
	}
}

func (d Decimal) String() string {
	if d.Exp == 0 {
		return fmt.Sprintf("%d.", d.Sig)
	}
	scale := int64(math.Pow10(int(d.Exp)))
	return fmt.Sprintf("%d.%d", d.Sig/scale, d.Sig%scale)
}
