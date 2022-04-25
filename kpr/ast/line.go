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

func (b *BadLine) Pos() token.Pos {
	return b.From
}

func (b *BadLine) End() token.Pos {
	return b.To
}

func (*BadLine) lineNode() {}

// A SplitLine node represents a split line node in a transaction.
type SplitLine struct {
	Account *BasicValue // STRING
	Amount  *Amount
}

func (s *SplitLine) Pos() token.Pos {
	return s.Account.Pos()
}

func (s *SplitLine) End() token.Pos {
	if s.Amount == nil {
		return s.Account.End()
	}
	return s.Amount.End()
}

func (*SplitLine) lineNode() {}

// An AmountLine node represents an amount line node.
type AmountLine struct {
	*Amount
}

func (*AmountLine) lineNode() {}

// A MetadataLine node represents a metadata line.
type MetadataLine struct {
	TokPos token.Pos
	Key    *BasicValue // STRING
	Val    *BasicValue // STRING
}

func (v *MetadataLine) Pos() token.Pos {
	return v.TokPos
}

func (v *MetadataLine) End() token.Pos {
	return v.Val.End()
}

func (*MetadataLine) lineNode() {}
