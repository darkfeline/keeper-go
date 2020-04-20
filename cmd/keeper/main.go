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
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "keeper: no command specified\n")
		os.Exit(2)
	}
	n := args[0]
	for _, c := range commands {
		if n == c.name {
			c.run(c, args[1:])
			return
		}
	}
	fmt.Fprintf(os.Stderr, "keeper: unknown command %s\n", n)
	os.Exit(2)
}

var commands = []*command{
	checkCmd,
}

type command struct {
	name string
	run  func(*command, []string)
}
