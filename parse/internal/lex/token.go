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

package lex

// Token is a lexed token.
type Token struct {
	Typ TokenType
	Val string
	Pos Pos
}

// go:generate stringer -type=TokenType

type TokenType uint8

const (
	// TokError is emitted for lex errors.  The error string is
	// stored in the Val field.
	TokError TokenType = iota
	// TokEOF is emitted when lexing terminates.
	TokEOF

	TokNewline
	TokDot
	TokDecimal
	TokDate
	TokKeyword
	TokAccount
	TokUnit
	TokString
)
