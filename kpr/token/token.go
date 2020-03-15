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

// Package token defines lexical tokens for parsing keeper files.
package token

// Token is the set of lexical tokens for keeper files.
type Token int

//go:generate stringer -type=Token

const (
	// Special
	ILLEGAL Token = iota
	EOF
	COMMENT

	// Syntactic
	NEWLINE

	// Values
	STRING
	UNIT_SYM
	ACCOUNT
	DECIMAL
	DATE

	// Keywords
	TX
	BALANCE
	UNIT
	END
)
