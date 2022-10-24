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

// A CompileArgs describes arguments to Compile.
type CompileArgs struct {
	Inputs []CompileInput
	// If set, only consider entries up to and including the
	// specified date.
	Ending civil.Date
}

// A CompileInput defines an input source for Compile.
type CompileInput interface {
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

// Bytes returns an option that specifies input bytes.
func Bytes(filename string, src []byte) CompileInput {
	return inputBytes{
		filename: filename,
		src:      src,
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

// File returns an option that specifies input files.
func File(filename ...string) []CompileInput {
	var i []CompileInput
	for _, f := range filename {
		i = append(i, inputFile{filename: f})
	}
	return i
}
