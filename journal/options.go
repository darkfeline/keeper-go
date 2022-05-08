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

package journal

import (
	"io/ioutil"
	"path/filepath"

	"cloud.google.com/go/civil"
)

// An Option is passed to Compile to configure compilation.
type Option interface {
	setOptions(*options)
}

type options struct {
	inputs []input
	ending civil.Date
}

func makeOptions(o []Option) options {
	var op options
	for _, o := range o {
		o.setOptions(&op)
	}
	return op
}

// An input defines an input source for compiling a journal.
type input interface {
	Filename() string
	Src() ([]byte, error)
}

type inputBytes struct {
	filename string
	src      []byte
}

func (o inputBytes) Filename() string {
	return o.filename
}

func (o inputBytes) Src() ([]byte, error) {
	return o.src, nil
}

func (o inputBytes) setOptions(opt *options) {
	opt.inputs = append(opt.inputs, o)
}

// Bytes returns an option that specifies input bytes.
func Bytes(filename string, src []byte) Option {
	return inputBytes{
		filename: filename,
		src:      src,
	}
}

type multiOpt []Option

func (o multiOpt) setOptions(opt *options) {
	for _, o := range o {
		o.setOptions(opt)
	}
}

type inputFile struct {
	filename string
}

func (o inputFile) Filename() string {
	return filepath.Base(o.filename)
}

func (o inputFile) Src() ([]byte, error) {
	src, err := ioutil.ReadFile(o.filename)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func (o inputFile) setOptions(opt *options) {
	opt.inputs = append(opt.inputs, o)
}

// File returns an option that specifies input files.
func File(filename ...string) Option {
	var o multiOpt
	for _, f := range filename {
		o = append(o, inputFile{filename: f})
	}
	return o
}

type endingOpt struct {
	date civil.Date
}

func (o endingOpt) setOptions(opt *options) {
	opt.ending = o.date
}

// Ending returns an option that limits a compiled journal to entries
// ending on or before the given date.
func Ending(d civil.Date) Option {
	return endingOpt{
		date: d,
	}
}
