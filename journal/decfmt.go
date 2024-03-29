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
	"math/big"
	"strings"
)

// decFormat does decimal formatting for n/scale.
// scale must be a positive multiple of 10.
func decFormat(n *big.Int, scale uint64) string {
	r := newRat()
	defer ratPool.Put(r)
	r2 := newRat()
	defer ratPool.Put(r2)
	r.Quo(r.SetInt(n), r2.SetUint64(scale))

	s := r.FloatString(log10(scale))
	digits := 0
count:
	for _, r := range s {
		switch r {
		case '-':
		case '.':
			break count
		default:
			digits++
		}
	}
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(r)
		if r == '-' || digits < 3 {
			continue
		}
		digits--
		if digits%3 == 0 {
			b.WriteRune(',')
		}
	}
	return b.String()
}

func log10(scale uint64) int {
	n := 0
	for ; scale > 1; scale /= 10 {
		n++
	}
	return n
}
