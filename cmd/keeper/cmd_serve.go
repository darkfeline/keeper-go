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
	"net/http"
	"os"

	"go.felesatra.moe/keeper/cmd/keeper/internal/webui"
	"go.felesatra.moe/keeper/journal"
)

var serveCmd = &command{
	usageLine: "serve [-port port]",
	run: func(cmd *command, args []string) {
		fs := flag.NewFlagSet(cmd.name(), flag.ExitOnError)
		port := fs.String("port", "8888", "Port to listen on")
		_ = fs.Parse(args) // cannot return error due to ExitOnError flag
		var o []journal.Option
		for _, f := range fs.Args() {
			o = append(o, journal.File(f))
		}

		fmt.Fprintf(os.Stderr, "Listening on port %s\n", *port)
		h := webui.NewHandler(o)
		log.Fatal(http.ListenAndServe(":"+*port, h))
	},
}
