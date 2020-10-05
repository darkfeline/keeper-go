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
	"log"
	"net"
	"net/http"
	"os"

	"github.com/coreos/go-systemd/activation"
	"go.felesatra.moe/keeper/cmd/keeper/internal/webui"
	"go.felesatra.moe/keeper/journal"
)

var serveCmd = &command{
	usageLine: "serve [-port port] [files]",
	run: func(cmd *command, args []string) {
		fs := flag.NewFlagSet(cmd.name(), flag.ExitOnError)
		port := fs.String("port", "8888", "Port to listen on")
		if err := fs.Parse(args); err != nil {
			panic(err)
		}
		var listener net.Listener
		ls, err := activation.Listeners()
		if err != nil {
			log.Fatal(err)
		}
		if len(ls) >= 1 {
			listener = ls[0]
			fmt.Fprintf(os.Stderr, "Using activation socket\n")
		} else {
			listener, err = net.Listen("tcp", *port)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Fprintf(os.Stderr, "Listening on port %s\n", *port)
		}
		o := []journal.Option{journal.File(fs.Args()...)}
		h := webui.NewHandler(o)
		log.Fatal(http.Serve(listener, h))
	},
}
