// Code generated by "binpack -name indexText index.html"; DO NOT EDIT.

package webui

const indexText = "{{define \"body\" -}}\n{{if .BalanceErrors -}}\n<p>Balance errors!</p>\n<table>\n  <thead>\n    <tr>\n      <th>Loc</th>\n      <th>Date</th>\n      <th>Account</th>\n      <th>Declared</th>\n      <th>Actual</th>\n      <th>Diff</th>\n    </tr>\n  </thead>\n  <tbody>\n    {{- range .BalanceErrors}}\n    <tr>\n      <td>{{.Pos}}</td>\n      <td>{{.Date}}</td>\n      <td><a href=\"/reconcile?account={{.Account}}\">{{.Account}}</a></td>\n      <td>{{.Declared}}</td>\n      <td>{{.Actual}}</td>\n      <td>{{.Diff}}</td>\n      <tr>\n    {{- end}}\n  </tbody>\n</table>\n{{end -}}\n<ul>\n  <li><a href=\"/trial\">Trial Balance</a></li>\n  <li><a href=\"/income\">Income</a></li>\n  <li><a href=\"/capital\">Capital</a></li>\n  <li><a href=\"/balance\">Balance Sheet</a></li>\n  <li><a href=\"/cash\">Cash Flow</a></li>\n</ul>\n{{- end}}\n"
