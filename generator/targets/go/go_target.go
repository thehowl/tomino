package gotarget

import (
	"fmt"
	"io"
	"text/template"

	"github.com/thehowl/tomino/generator/ir"

	_ "embed"
)

//go:embed template.tmpl
var templateSource string

var tpl = template.Must(template.New("").
	Funcs(template.FuncMap{
		"throw": func(s string, args ...any) (string, error) {
			return "", fmt.Errorf(s, args...)
		},
	}).
	Parse(templateSource))

func Write(w io.Writer, messages []ir.StructRecord) error {
	return tpl.ExecuteTemplate(w, "main", messages)
}
