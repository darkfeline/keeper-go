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

	"go.felesatra.moe/keeper/parse/raw"
)

type commonEntry struct {
	e interface{}
}

func (c commonEntry) summary() string {
	switch e := c.e.(type) {
	case raw.BalanceEntry:
		return fmt.Sprintf("balance %v %v", e.Date, e.Account)
	case raw.UnitEntry:
		return fmt.Sprintf("unit %v", e.Symbol)
	case raw.TransactionEntry:
		return fmt.Sprintf("transaction %v %#v", e.Date, e.Description)
	default:
		panic(e)
	}
}

func (c commonEntry) lineno() int {
	switch e := c.e.(type) {
	case raw.BalanceEntry:
		return e.Line
	case raw.UnitEntry:
		return e.Line
	case raw.TransactionEntry:
		return e.Line
	default:
		panic(e)
	}
}

func processErr(e interface{}, v interface{}) error {
	return processErrf(e, "%v", v)
}

func processErrf(e interface{}, format string, v ...interface{}) error {
	c := commonEntry{e: e}
	msg := fmt.Sprintf(format, v...)
	return fmt.Errorf("entry %v (line %d): %v", c.summary(), c.lineno(), msg)
}

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
	return fmt.Sprintf("%d errors while processing:\n  - %v",
		len(e.errs),
		strings.Join(s, "\n  - "))
}
