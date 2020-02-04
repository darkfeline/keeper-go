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

package ast

import (
	"go.felesatra.moe/keeper/kpr/token"
)

// Entries

type Entry interface {
	Node
	entry()
}

type BadEntry struct {
	From, To token.Pos
}

func (b BadEntry) Pos() token.Pos {
	return b.From
}

func (b BadEntry) End() token.Pos {
	return b.To
}

func (BadEntry) entry() {}

type BalanceHeader struct {
	TokPos  token.Pos
	Date    BasicValue // DATE
	Account BasicValue // ACCOUNT
}

func (b BalanceHeader) Pos() token.Pos {
	return b.TokPos
}

func (b BalanceHeader) End() token.Pos {
	return b.Account.End()
}

type SingleBalance struct {
	BalanceHeader
	Amount Amount
}

func (b SingleBalance) End() token.Pos {
	return b.Amount.End()
}

func (SingleBalance) entry() {}

type MultiBalance struct {
	BalanceHeader
	Amounts []LineNode // AmountLine, BadLine
	Dot     Dot
}

func (b MultiBalance) End() token.Pos {
	return b.Dot.End()
}

func (MultiBalance) entry() {}

type UnitDecl struct {
	TokPos token.Pos
	Unit   BasicValue // IDENT
	Scale  BasicValue // DECIMAL
}

func (u UnitDecl) Pos() token.Pos {
	return u.TokPos
}

func (u UnitDecl) End() token.Pos {
	return u.Scale.End()
}

func (UnitDecl) entry() {}

type Transaction struct {
	TokPos      token.Pos
	Date        BasicValue // DATE
	Description BasicValue // STRING
	Splits      []LineNode // Split, BadLine
	Dot         Dot
}

func (t Transaction) Pos() token.Pos {
	return t.TokPos
}

func (t Transaction) End() token.Pos {
	return t.Dot.End()
}

func (Transaction) entry() {}

// Line nodes

type LineNode interface {
	Node
	lineNode()
}

type BadLine struct {
	From, To token.Pos
}

func (b BadLine) Pos() token.Pos {
	return b.From
}

func (b BadLine) End() token.Pos {
	return b.To
}

func (BadLine) lineNode() {}

type Split struct {
	Account BasicValue // STRING
	Amount  *Amount
}

func (s Split) Pos() token.Pos {
	return s.Account.Pos()
}

func (s Split) End() token.Pos {
	if s.Amount == nil {
		return s.Account.End()
	}
	return s.Amount.End()
}

func (Split) lineNode() {}

type AmountLine struct {
	Amount
}

func (AmountLine) lineNode() {}

// Simple nodes

type Node interface {
	Pos() token.Pos
	End() token.Pos
}

type Dot struct {
	TokPos token.Pos
}

func (d Dot) Pos() token.Pos {
	return d.TokPos
}

func (d Dot) End() token.Pos {
	return token.Pos(int(d.TokPos) + 1)
}

type Amount struct {
	Decimal BasicValue // DECIMAL
	Unit    BasicValue // IDENT
}

func (a Amount) Pos() token.Pos {
	return a.Decimal.Pos()
}

func (a Amount) End() token.Pos {
	return a.Unit.End()
}

type BasicValue struct {
	ValuePos token.Pos
	Kind     token.Token // STRING, IDENT, ACCOUNT, DECIMAL, DATE
	Value    string
}

func (v BasicValue) Pos() token.Pos {
	return v.ValuePos
}

func (v BasicValue) End() token.Pos {
	return token.Pos(int(v.ValuePos) + len(v.Value))
}
