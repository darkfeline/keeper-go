// Code generated by "binpack -name baseText base.html"; DO NOT EDIT.

package templates

const baseText = "<!DOCTYPE HTML>\n<html lang=\"en\">\n  <head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    <link rel=\"stylesheet\" type=\"text/css\" href=\"/style.css\">\n    <title>Keeper Web UI{{if .Title}} - {{.Title}}{{end}}</title>\n  </head>\n  <body>\n    <header>\n      <nav class=\"sitenav\">\n        <h1><a href=\"/\">Keeper</a></h1>\n        <ul>\n          <li><a href=\"/trial\">Trial Balance</a></li>\n          <li><a href=\"/income\">Income</a></li>\n          <li><a href=\"/capital\">Capital</a></li>\n          <li><a href=\"/balance\">Balance Sheet</a></li>\n          <li><a href=\"/cash\">Cash Flow</a></li>\n        </ul>\n      </nav>\n    </header>\n    {{block \"body\" .}}{{.Body}}{{end}}\n  </body>\n</html>\n"