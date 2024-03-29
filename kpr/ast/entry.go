// Copyright (C) 2020  Allen Li
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

import "go.felesatra.moe/keeper/kpr/token"

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

func (b *BadEntry) Pos() token.Pos {
	return b.From
}

func (b *BadEntry) End() token.Pos {
	return b.To
}

func (*BadEntry) entry() {}

// A SingleBalance node represents a balance entry node on a single
// line.
type SingleBalance struct {
	BalanceHeader
	Amount *Amount
}

func (b *SingleBalance) End() token.Pos {
	return b.Amount.End()
}

func (*SingleBalance) entry() {}

// A MultiBalance node represents a balance entry node spanning
// multiple lines.
type MultiBalance struct {
	BalanceHeader
	Amounts []LineNode // AmountLine, BadLine
	EndTok  *End
}

func (b *MultiBalance) End() token.Pos {
	return b.EndTok.End()
}

func (*MultiBalance) entry() {}

// A BalanceHeader contains the fields shared by balance entry nodes.
type BalanceHeader struct {
	TokPos  token.Pos
	Token   token.Token
	Date    *BasicValue // DATE
	Account *BasicValue // ACCTNAME
}

func (b *BalanceHeader) Pos() token.Pos {
	return b.TokPos
}

func (b *BalanceHeader) End() token.Pos {
	return b.Account.End()
}

// A UnitDecl node represents a unit declaration entry node.
type UnitDecl struct {
	TokPos token.Pos
	Unit   *BasicValue // USYMBOL
	Scale  *BasicValue // DECIMAL
}

func (u *UnitDecl) Pos() token.Pos {
	return u.TokPos
}

func (u *UnitDecl) End() token.Pos {
	return u.Scale.End()
}

func (*UnitDecl) entry() {}

// A Transaction node represents a transaction entry node.
type Transaction struct {
	TokPos      token.Pos
	Date        *BasicValue // DATE
	Description *BasicValue // STRING
	Splits      []LineNode  // SplitLine, BadLine
	EndTok      *End
}

func (t *Transaction) Pos() token.Pos {
	return t.TokPos
}

func (t *Transaction) End() token.Pos {
	return t.EndTok.End()
}

func (*Transaction) entry() {}

// A DisableAccount node represents a disable account entry node.
type DisableAccount struct {
	TokPos  token.Pos
	Date    *BasicValue // DATE
	Account *BasicValue // ACCTNAME
}

func (c *DisableAccount) Pos() token.Pos {
	return c.TokPos
}

func (c *DisableAccount) End() token.Pos {
	return c.Account.End()
}

func (*DisableAccount) entry() {}

// A DeclareAccount node represents a declare account entry node.
type DeclareAccount struct {
	TokPos   token.Pos
	Account  *BasicValue // ACCTNAME
	Metadata []LineNode  // MetadataLine, BadLine
	EndTok   *End
}

func (c *DeclareAccount) Pos() token.Pos {
	return c.TokPos
}

func (c *DeclareAccount) End() token.Pos {
	return c.EndTok.End()
}

func (*DeclareAccount) entry() {}
