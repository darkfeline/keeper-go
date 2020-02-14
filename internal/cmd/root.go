// Copyright (C) 2019  Allen Li
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

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"cloud.google.com/go/civil"
	"github.com/spf13/cobra"
	"go.felesatra.moe/keeper/book"
	"go.felesatra.moe/keeper/cmd/internal/colfmt"
)

var rootCmd = &cobra.Command{
	Use:          "keeper",
	Short:        "keeper is plain text accounting software",
	SilenceUsage: true,
}

var (
	format       string
	startDateStr string
	endDateStr   string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&format, "format", prettyFmt, "output format")
	rootCmd.PersistentFlags().StringVar(&startDateStr, "start", "", "start date")
	rootCmd.PersistentFlags().StringVar(&endDateStr, "end", "", "end date")
}

const (
	prettyFmt = "pretty"
	tabFmt    = "tab"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func compileFile(path string) (*book.Book, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return compile(src)
}

func compile(src []byte) (*book.Book, error) {
	var o []book.Option
	d, err := startDate()
	if err != nil {
		return nil, err
	}
	if d.IsValid() {
		o = append(o, book.Starting(d))
	}
	d, err = endDate()
	if err != nil {
		return nil, err
	}
	if d.IsValid() {
		o = append(o, book.Ending(d))
	}
	return book.Compile(src, o...)
}

func startDate() (civil.Date, error) {
	if startDateStr == "" {
		return civil.Date{}, nil
	}
	d, err := civil.ParseDate(startDateStr)
	if err != nil {
		return civil.Date{}, err
	}
	return d, nil
}

func endDate() (civil.Date, error) {
	if endDateStr == "" {
		return civil.Date{}, nil
	}
	d, err := civil.ParseDate(endDateStr)
	if err != nil {
		return civil.Date{}, err
	}
	return d, nil
}

type formatter func(io.Writer, interface{}) error

func getFormatter(format string) (formatter, error) {
	switch format {
	case tabFmt:
		return colfmt.FormatTab, nil
	case prettyFmt:
		return colfmt.Format, nil
	default:
		return nil, fmt.Errorf("unknown format %v", format)
	}
}
