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

package main

import (
	"flag"
	"fmt"
	"os"

	"go.felesatra.moe/keeper/internal/account"
)

// A configFlag is a flag.Value for loading a account.Config via a flag.
type configFlag struct {
	path string
	c    *account.Classifier
}

func (f *configFlag) Set(s string) error {
	f.path = s
	fi, err := os.Open(s)
	if err != nil {
		return fmt.Errorf("set config flag: %s", err)
	}
	if err := account.LoadClassifier(f.c, fi); err != nil {
		return fmt.Errorf("set config flag: %s", err)
	}
	return nil
}

func (f *configFlag) String() string {
	return f.path
}

// configPath adds a flag for the path to a account config file and
// returns a account.Config that is modified when flags are parse.
func configPath(fs *flag.FlagSet) *account.Classifier {
	v := &configFlag{
		c: &account.Classifier{},
	}
	fs.Var(v, "config", "Path to account config file")
	return v.c
}
