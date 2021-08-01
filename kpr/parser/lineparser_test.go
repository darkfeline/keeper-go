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
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.felesatra.moe/keeper/kpr/ast"
	"go.felesatra.moe/keeper/kpr/scanner"
	"go.felesatra.moe/keeper/kpr/token"
)

func TestLineParser(t *testing.T) {
	t.Parallel()
	cases := []struct {
		desc  string
		input string
		mode  scanner.Mode
		want  []*line
	}{
		{
			desc: "lines",
			input: `tx 2020-01-02 "blah"
end
`,
			want: []*line{
				{
					tokens: []tokenInfo{
						{1, token.TX, "tx"},
						{4, token.DATE, "2020-01-02"},
						{15, token.STRING, `"blah"`},
					},
					eol: tokenInfo{21, token.NEWLINE, "\n"},
				},
				{
					tokens: []tokenInfo{
						{22, token.END, "end"},
					},
					eol: tokenInfo{25, token.NEWLINE, "\n"},
				},
				{
					tokens: []tokenInfo{},
					eol:    tokenInfo{26, token.EOF, ""},
				},
			},
		},
		{
			desc:  "comment only",
			input: `# blah`,
			mode:  scanner.ScanComments,
			want: []*line{
				{
					tokens:  []tokenInfo{},
					comment: &ast.Comment{TokPos: 1, Text: "# blah"},
					eol:     tokenInfo{7, token.EOF, ""},
				},
			},
		},
		{
			desc:  "skip comment",
			input: `# blah`,
			want: []*line{
				{
					tokens: []tokenInfo{},
					eol:    tokenInfo{7, token.EOF, ""},
				},
			},
		},
	}
	o := cmp.AllowUnexported(tokenInfo{}, line{})
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			got, err := parseTestLines([]byte(c.input), c.mode)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(c.want, got, o); diff != "" {
				t.Errorf("lines mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func parseTestLines(b []byte, m scanner.Mode) ([]*line, error) {
	fset := token.NewFileSet()
	f := fset.AddFile("", -1, len(b))
	var errs scanner.ErrorList
	p := newLineParser(f, b, errs.Add, m)
	return p.parseLines(), errs.Err()
}
