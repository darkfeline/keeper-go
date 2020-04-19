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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type decimal struct {
	number int64
	// Scale indicates the minimum fractional unit amount,
	// e.g. 100 means 0.01 is the smallest amount.
	// Should be a multiple of 10.
	scale int64
}

func parseDecimal(s string) (decimal, error) {
	if len(s) == 0 {
		return decimal{}, errors.New("parse decimal: empty string")
	}
	neg := s[0] == '-'
	s = strings.Replace(s, ",", "", -1)
	split := len(s)
	for i, b := range s {
		if b == '.' {
			split = i
			break
		}
	}
	before, err := strconv.ParseInt(s[:split], 10, 64)
	if err != nil {
		return decimal{}, fmt.Errorf("parse decimal %#v: %s", s, err)
	}
	var after int64
	if split+1 < len(s) {
		after, err = strconv.ParseInt(s[split+1:], 10, 64)
		if err != nil {
			return decimal{}, fmt.Errorf("parse decimal %#v: %s", s, err)
		}
	}
	if neg {
		after = -after
	}
	split = len(s) - split
	if split > 0 {
		split--
	}
	scale := int64(math.Pow10(split))
	return decimal{
		number: before*scale + after,
		scale:  scale,
	}, nil
}
