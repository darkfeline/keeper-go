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

import "cloud.google.com/go/civil"

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
		o.(optionSetter)(&op)
	}
	return op
}

type input interface {
	input()
}

type inputBytes struct {
	filename string
	src      []byte
}

func (inputBytes) input() {}

// Bytes returns an option that specifies input bytes.
func Bytes(filename string, src []byte) Option {
	return optionSetter(func(o *options) {
		o.inputs = append(o.inputs, inputBytes{
			filename: filename,
			src:      src,
		})
	})
}

type inputFile struct {
	filename string
}

func (inputFile) input() {}

// File returns an option that specifies input files.
func File(filename ...string) Option {
	return optionSetter(func(o *options) {
		for _, f := range filename {
			o.inputs = append(o.inputs, inputFile{filename: f})
		}
	})
}

// Ending returns an option that limits a compiled journal to entries
// ending on or before the given date.
func Ending(d civil.Date) Option {
	return optionSetter(func(o *options) {
		o.ending = d
	})
}

// An optionSetter is a function that implements the Option interface.
type optionSetter func(*options)

func (f optionSetter) setOptions(o *options) {
	f(o)
}
