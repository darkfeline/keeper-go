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

package webui

import (
	"bytes"
	"net/http"

	"go.felesatra.moe/keeper/journal"
)

func NewHandler(o []journal.Option) http.Handler {
	h := handler{o}
	m := http.NewServeMux()
	m.HandleFunc("/", h.handleIndex)
	m.HandleFunc("/reconcile", h.handleReconcile)
	m.HandleFunc("/trial", h.handleTrial)
	return m
}

type handler struct {
	o []journal.Option
}

func (h handler) handleIndex(w http.ResponseWriter, req *http.Request) {
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var b bytes.Buffer
	d := indexData{
		BalanceErrors: j.BalanceErrors,
	}
	if err := indexTemplate.Execute(&b, d); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(b.Bytes())
}

func (h handler) handleReconcile(w http.ResponseWriter, req *http.Request) {
	panic("Not implemented")
}

func (h handler) handleTrial(w http.ResponseWriter, req *http.Request) {
	panic("Not implemented")
	j, err := h.compile()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = j
	// XXXXXXXXXXXXXXXXXXXXX
	var b bytes.Buffer
	if err := indexTemplate.Execute(&b, nil); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(b.Bytes())
}

func (h handler) compile() (*journal.Journal, error) {
	return journal.Compile(h.o...)
}
