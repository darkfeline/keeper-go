// Copyright (C) 2021  Allen Li
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

package parser

import (
	"fmt"

	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/scanner"
	"go.felesatra.moe/keeper/kpr/token"
)

// A lineParser parses lines in a keeper file.
type lineParser struct {
	f    *token.File
	s    scanner.Scanner
	errs scanner.ErrorList
}

func newLineParser(fset *token.FileSet, filename string, src []byte, m scanner.Mode) *lineParser {
	p := &lineParser{
		f: fset.AddFile(filename, -1, len(src)),
	}
	p.s.Init(p.f, src, p.errs.Add, m)
	return p
}

// Parse all lines.
func (p *lineParser) parseLines() []line {
	var lines []line
	for {
		l := p.parseLine()
		lines = append(lines, l)
		if l.eol.tok == token.EOF {
			return lines
		}
	}
}

// Parse a single line.
func (p *lineParser) parseLine() line {
	l := line{
		tokens: make([]tokenInfo, 0, 4),
	}
	for {
		pos, tok, lit := p.s.Scan()
		switch tok {
		case token.NEWLINE, token.EOF:
			l.eol = tokenInfo{pos: pos, tok: tok, lit: lit}
			return l
		case token.COMMENT:
			l.comment = &ast.Comment{
				TokPos: pos,
				Text:   lit,
			}
		default:
			if l.comment != nil {
				panic(fmt.Sprintf("unexpected %v %v %v", pos, tok, lit))
			}
			l.tokens = append(l.tokens, tokenInfo{pos: pos, tok: tok, lit: lit})
		}
	}
}

type line struct {
	tokens []tokenInfo
	// Comment for the line, if any.
	comment *ast.Comment
	// Token that ends the line, either NEWLINE or EOF.
	eol tokenInfo
}

type tokenInfo struct {
	pos token.Pos
	tok token.Token
	lit string
}
