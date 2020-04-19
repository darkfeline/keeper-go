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
	"strings"
)

// decFormat does decimal formatting for n/scale.
// scale must be a positive multiple of 10.
func decFormat(n int64, scale int64) string {
	var b strings.Builder
	if n < 0 {
		b.WriteRune('-')
		n = -n
	}
	logScale := log10(scale)
	d := fmt.Sprintf("%0*d", logScale+1, n)
	split := len(d) - logScale
	before, after := d[:split], d[split:]
	for i, r := range before {
		b.WriteRune(r)
		if i := len(before) - i; i > 1 && i%3 == 1 {
			b.WriteRune(',')
		}
	}
	if len(after) == 0 {
		return b.String()
	}
	b.WriteRune('.')
	b.WriteString(after)
	return b.String()
}

func log10(scale int64) int {
	n := 0
	for ; scale > 1; scale /= 10 {
		n++
	}
	return n
}
