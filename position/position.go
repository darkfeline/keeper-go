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

package position

import (
	"fmt"

	"go.felesatra.moe/keeper/fixed"
)

type Position struct {
	Amount fixed.Fixed
	Unit   Unit
}

type Unit string

func (p Position) String() string {
	return fmt.Sprintf("%v %v", p.Amount, p.Unit)
}

func Add(ps []Position, f fixed.Fixed, u Unit) []Position {
	for i, p := range ps {
		if p.Unit == u {
			ps[i].Amount = p.Amount.Add(f)
			return ps
		}
	}
	return append(ps, Position{Amount: f, Unit: u})
}

func Merge(p []Position, v []Position) []Position {
	for _, v := range v {
		p = Add(p, v.Amount, v.Unit)
	}
	return p
}
