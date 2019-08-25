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

package stage1

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"go.felesatra.moe/keeper"
	"golang.org/x/xerrors"
)

func Parse(r io.Reader) ([]Transaction, error) {
	var ts []Transaction
	s := bufio.NewScanner(r)
	linum := 0
	for s.Scan() {
		line := s.Text()
		t, err := parseTransaction(line)
		if err != nil {
			return nil, xerrors.Errorf("parse line %d: %s", linum, err)
		}
		ts = append(ts, t)
		linum++
	}
	if err := s.Err(); err != nil {
		return nil, xerrors.Errorf("parse transactions: %w", err)
	}
	return ts, nil
}

func parseTransaction(line string) (Transaction, error) {
	parts := strings.Fields(line)
	if len(parts) != 4 {
		return Transaction{}, fmt.Errorf("parse transaction %#v: invalid", line)
	}
	a1, err := ParseAccount(parts[0])
	if err != nil {
		return Transaction{}, xerrors.Errorf("parse transaction %#v: %s", line, err)
	}
	a2, err := ParseAccount(parts[1])
	if err != nil {
		return Transaction{}, xerrors.Errorf("parse transaction %#v: %s", line, err)
	}
	d, err := keeper.ParseFixed(parts[2])
	if err != nil {
		return Transaction{}, xerrors.Errorf("parse transaction %#v: %s", line, err)
	}
	return Transaction{
		From:   a1,
		To:     a2,
		Amount: d,
		Unit:   keeper.Unit(parts[3]),
	}, nil
}

func WriteTransactions(w io.Writer, ts []Transaction) error {
	bw := bufio.NewWriter(w)
	for _, t := range ts {
		fmt.Fprintf(bw, "%s %s %s %s", t.From, t.To, t.Amount, t.Unit)
	}
	if err := bw.Flush(); err != nil {
		return xerrors.Errorf("write transactions: %w", err)
	}
	return nil
}
