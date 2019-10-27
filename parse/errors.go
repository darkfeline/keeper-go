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

package parse

import (
	"fmt"
	"strings"
)

// processError is returned for errors processing parsed entries.
type processError struct {
	errs []error
}

func (e processError) Error() string {
	n := len(e.errs)
	if n == 0 {
		return "error while processing"
	}
	s := make([]string, n)
	for i, e := range e.errs {
		s[i] = e.Error()
	}
	return fmt.Sprintf("%d errors while processing:\n  -%v",
		len(e.errs),
		strings.Join(s, "\n  -"))
}
