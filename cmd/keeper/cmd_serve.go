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
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/cmd/keeper/internal/webui"
	"go.felesatra.moe/keeper/journal"
)

func init() {
	serveCmd.Flags().StringVar(&servePort, "port", "8888", "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}

var servePort string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run web UI",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var o []journal.Option
		for _, f := range args {
			o = append(o, journal.File(f))
		}

		fmt.Fprintf(os.Stderr, "Listening on port %s\n", servePort)
		h := webui.NewHandler(o)
		log.Fatal(http.ListenAndServe(":"+servePort, h))
	},
}
