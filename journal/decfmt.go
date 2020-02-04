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
)

// decFormat does decimal formatting for n/scale.
// scale must be a positive multiple of 10.
func decFormat(n int64, scale int64) string {
	if scale <= 1 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d."+fracFmtStr(scale), n/scale, abs(n%scale))
}

func fracFmtStr(scale int64) string {
	n := 0
	for ; scale > 1; scale /= 10 {
		n++
	}
	return fmt.Sprintf("%%0%dd", n)
}

func abs(x int64) int64 {
	y := x >> 63
	return (x ^ y) - y
}
