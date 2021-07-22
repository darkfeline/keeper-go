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
// kpr files.
package ast

import (
	"go.felesatra.moe/keeper/kpr/token"
)

// A File represents a source file.
type File struct {
	Entries  []Entry
	Comments []*CommentGroup
}

// All node types implement the Node interface.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// An End node represents an end keyword ending a multiple line entry.
type End struct {
	TokPos token.Pos
}

func (d *End) Pos() token.Pos {
	return d.TokPos
}

func (d *End) End() token.Pos {
	return token.Pos(int(d.TokPos) + len("end"))
}

// An Amount node represents an amount.
type Amount struct {
	Decimal *BasicValue // DECIMAL
	Unit    *BasicValue // USYMBOL
}

func (a *Amount) Pos() token.Pos {
	return a.Decimal.Pos()
}

func (a *Amount) End() token.Pos {
	return a.Unit.End()
}

// A BasicValue node represents a basic single token value.
type BasicValue struct {
	ValuePos token.Pos
	Kind     token.Token // STRING, USYMBOL, ACCTNAME, DECIMAL, DATE
	Value    string
}

func (v *BasicValue) Pos() token.Pos {
	return v.ValuePos
}

func (v *BasicValue) End() token.Pos {
	return token.Pos(int(v.ValuePos) + len(v.Value))
}

// A Comment node represents a comment.
type Comment struct {
	// Position of starting #
	TokPos token.Pos
	// Comment text, including # and excluding trailing newline.
	Text string
}

func (c *Comment) Pos() token.Pos {
	return c.TokPos
}

func (c *Comment) End() token.Pos {
	return token.Pos(int(c.TokPos) + len(c.Text))
}

// A CommentGroup represents consecutive comments.
type CommentGroup struct {
	// Must be non-empty
	List []*Comment
}

func (c *CommentGroup) Pos() token.Pos {
	return c.List[0].Pos()
}

func (c *CommentGroup) End() token.Pos {
	return c.List[len(c.List)-1].End()
}
