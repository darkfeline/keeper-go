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

// Package ast declares the types used to represent syntax trees for
// keeper files.
package ast

import (
	"go.felesatra.moe/keeper/kpr/token"
)

// All entry nodes implement Entry.
type Entry interface {
	Node
	entry()
}

// A BadEntry node is a placeholder for an entry containing syntax
// errors for which a correct entry node cannot be created.
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

// A BalanceHeader contains the fields shared by balance entry nodes.
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

// A SingleBalance node represents a balance entry node on a single
// line.
type SingleBalance struct {
	BalanceHeader
	Amount Amount
}

func (b SingleBalance) End() token.Pos {
	return b.Amount.End()
}

func (SingleBalance) entry() {}

// A MultiBalance node represents a balance entry node spanning
// multiple lines.
type MultiBalance struct {
	BalanceHeader
	Amounts []LineNode // AmountLine, BadLine
	Dot     Dot
}

func (b MultiBalance) End() token.Pos {
	return b.Dot.End()
}

func (MultiBalance) entry() {}

// An UnitDecl node represents a unit declaration entry node.
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

// A Transaction node represents a transaction entry node.
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

// All line nodes implement LineNode.
type LineNode interface {
	Node
	lineNode()
}

// A BadLine node is a placeholder for a line containing syntax
// errors for which a correct line node cannot be created.
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

// A Split node represents a split line node in a transaction.
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

// An AmountLine node represents an amount line node.
type AmountLine struct {
	Amount
}

func (AmountLine) lineNode() {}

// All node types implement the Node interface.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// A Dot node represents a dot ending a multiple line entry.
type Dot struct {
	TokPos token.Pos
}

func (d Dot) Pos() token.Pos {
	return d.TokPos
}

func (d Dot) End() token.Pos {
	return token.Pos(int(d.TokPos) + 1)
}

// An Amount node represents an amount.
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

// A BasicValue node represents a basic single token value.
type BasicValue struct {
	ValuePos token.Pos
	Kind     token.Token // STRING, UNIT_SYM, ACCOUNT, DECIMAL, DATE
	Value    string
}

func (v BasicValue) Pos() token.Pos {
	return v.ValuePos
}

func (v BasicValue) End() token.Pos {
	return token.Pos(int(v.ValuePos) + len(v.Value))
}
