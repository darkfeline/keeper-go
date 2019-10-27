package raw

import (
	"testing"

	"go.felesatra.moe/keeper/parse/internal/lex"
)

func TestParseStringTok(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{`"foo\\\n"`, `foo\n`},
		{`"foo\\""`, `foo\"`},
	}
	for _, c := range cases {
		c := c
		t.Run(c.input, func(t *testing.T) {
			t.Parallel()
			got, err := parseStringTok(lex.Token{
				Val: c.input,
				Typ: lex.TokString,
			})
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if got != c.want {
				t.Errorf("parse %#v = %#v; want %#v", c.input, got, c.want)
			}
		})
	}
}
