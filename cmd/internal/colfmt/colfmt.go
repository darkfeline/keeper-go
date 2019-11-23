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

package colfmt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func Format(w io.Writer, v interface{}) error {
	if err := checkType(v); err != nil {
		return fmt.Errorf("colspec: %v", err)
	}
	rv := reflect.ValueOf(v)
	n := rv.Len()
	bw := bufio.NewWriter(w)
	format := formatString(getSliceColspecs(v))
	for i := 0; i < n; i++ {
		f := structFields(rv.Index(i))
		fmt.Fprintf(bw, format, f...)
	}
	return bw.Flush()
}

func FormatTab(w io.Writer, v interface{}) error {
	if err := checkType(v); err != nil {
		return fmt.Errorf("colspec: %v", err)
	}
	rv := reflect.ValueOf(v)
	n := rv.Len()
	bw := bufio.NewWriter(w)
	format := tabFormatString(rv.Elem().NumField())
	for i := 0; i < n; i++ {
		f := structFields(rv.Index(i))
		fmt.Fprintf(bw, format, f...)
	}
	return bw.Flush()
}

func checkType(v interface{}) error {
	t := reflect.TypeOf(v)
	if k := t.Kind(); k != reflect.Slice {
		return errors.New("not a slice of structs of strings")
	}
	t = t.Elem()
	if k := t.Kind(); k != reflect.Struct {
		return errors.New("not a slice of structs of strings")
	}
	n := t.NumField()
	for i := 0; i < n; i++ {
		if k := t.Field(i).Type.Kind(); k != reflect.String {
			return errors.New("not a slice of structs of strings")
		}
	}
	return nil
}

func getSliceColspecs(v interface{}) []colspec {
	t := reflect.TypeOf(v)
	c := getStructColspecs(t.Elem())
	rv := reflect.ValueOf(v)
	n := rv.Len()
	for i := 0; i < n; i++ {
		updateColspecWidth(c, rv.Index(i))
	}
	return c
}

func getStructColspecs(t reflect.Type) []colspec {
	c := make([]colspec, t.NumField())
	for i := 0; i < len(c); i++ {
		f := t.Field(i)
		c[i] = parseTag(f.Tag.Get("colfmt"))
	}
	return c
}

func updateColspecWidth(c []colspec, v reflect.Value) {
	for i := 0; i < len(c); i++ {
		if i == len(c)-1 && c[i].Align == alignLeft {
			return
		}
		n := v.Field(i).Len()
		if c[i].Width < n {
			c[i].Width = n
		}
	}
}

func structFields(v reflect.Value) []interface{} {
	var f []interface{}
	n := v.NumField()
	for i := 0; i < n; i++ {
		f = append(f, v.Field(i))
	}
	return f
}

func parseTag(s string) colspec {
	parts := strings.Split(s, ",")
	var c colspec
	for _, p := range parts {
		switch p {
		case "right":
			c.Align = alignRight
		}
	}
	return c
}

type colspec struct {
	Width int
	Align alignment
}

func (c colspec) format() string {
	var b strings.Builder
	b.WriteByte('%')
	if c.Width > 0 {
		if c.Align == alignLeft {
			b.WriteByte('-')
		}
		fmt.Fprintf(&b, "%d", c.Width)
	}
	b.WriteByte('s')
	return b.String()
}

func formatString(c []colspec) string {
	switch len(c) {
	case 0:
		return ""
	case 1:
		return c[0].format()
	}
	var b strings.Builder
	b.WriteString(c[0].format())
	for _, c := range c[1:] {
		b.WriteByte(' ')
		b.WriteString(c.format())
	}
	b.WriteByte('\n')
	return b.String()
}

func tabFormatString(n int) string {
	p := make([]string, n)
	for i := range p {
		p[i] = "%s"
	}
	return strings.Join(p, "\t") + "\n"
}

type alignment int

const (
	alignLeft alignment = iota
	alignRight
)
