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
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		errf("no command specified")
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	for _, c := range commands {
		if cmd == c.name() {
			c.run(c, args)
			os.Exit(0)
		}
	}
	errf("unknown command %s", cmd)
	os.Exit(2)
}

var commands = []*command{
	checkCmd,
	closeCmd,
	serveCmd,
}

type command struct {
	usageLine string
	run       func(*command, []string)
}

func (c *command) name() string {
	return strings.SplitN(c.usageLine, " ", 2)[0]
}

func (c *command) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(c.name(), flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: keeper %s\n", c.usageLine)
		fs.PrintDefaults()
	}
	return fs
}

// errf prints annotated lines to stderr.
func errf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "keeper: "+format+"\n", v...)
}
