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
	"html/template"

	"go.felesatra.moe/keeper/journal"
)

//go:generate binpack -name baseText base.html

var baseTemplate = template.Must(template.New("base").Parse(baseText))

type baseData struct {
	Title string
	Body  template.HTML
}

//go:generate binpack -name indexText index.html

var indexTemplate = template.Must(baseTemplate.Parse(indexText))

type indexData struct {
	BalanceErrors []journal.BalanceAssert
}
