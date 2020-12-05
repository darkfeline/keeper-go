// Code generated by "binpack -name indexText index.html"; DO NOT EDIT.

package templates

const indexText = "{{- define \"body\" -}}\n{{if .BalanceErrors -}}\n<p>Balance errors!</p>\n<table>\n  <thead>\n    <tr>\n      <th>Ref</th>\n      <th>Date</th>\n      <th>Account</th>\n      <th>Declared</th>\n      <th>Actual</th>\n      <th>Diff</th>\n    </tr>\n  </thead>\n  <tbody>\n    {{- range .BalanceErrors}}\n    <tr>\n      <td>{{.Position}}</td>\n      <td>{{.Date}}</td>\n      <td><a href=\"/ledger?account={{.Account}}\">{{.Account}}</a></td>\n      <td class=\"amount\">{{.Declared}}</td>\n      <td class=\"amount\">{{.Actual}}</td>\n      <td class=\"amount\">{{.Diff}}</td>\n      <tr>\n    {{- end}}\n  </tbody>\n</table>\n{{else -}}\n<p>All OK</p>\n{{end -}}\n{{- end}}\n"
