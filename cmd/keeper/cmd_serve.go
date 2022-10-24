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
	"log"
	"net"
	"net/http"

	"go.felesatra.moe/keeper/internal/activation"
	"go.felesatra.moe/keeper/internal/webui"
	"go.felesatra.moe/keeper/journal"
)

var serveCmd = &command{
	usageLine: "serve [-addr address] [-config path] [files]",
	run: func(cmd *command, args []string) {
		fs := cmd.flagSet()
		c := fs.String("config", "", "Path to account config file")
		addr := fs.String("addr", "localhost:8888", "Address to listen on")
		fs.Parse(args)
		var listener net.Listener
		ls, err := activation.Listeners()
		if err != nil {
			log.Fatal(err)
		}
		if len(ls) >= 1 {
			listener = ls[0]
			log.Printf("using activation socket")
		} else {
			listener, err = net.Listen("tcp", *addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("listening on %s", *addr)
		}
		h := webui.NewHandler(*c, &journal.CompileArgs{
			Inputs: journal.File(fs.Args()...),
		})
		log.Fatal(http.Serve(listener, h))
	},
}
