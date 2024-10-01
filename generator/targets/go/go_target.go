package gotarget

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"text/template"

	"github.com/thehowl/tomino/generator/ir"

	_ "embed"
)

//go:embed template.tmpl
var templateSource string

var tUint64 = reflect.TypeOf(uint64(0))

var tpl = template.Must(template.New("template.tmpl").
	Funcs(template.FuncMap{
		"throw": func(s string, args ...any) (string, error) {
			return "", fmt.Errorf(s, args...)
		},
		"uvarint": func(i any) []byte {
			n := reflect.ValueOf(i).Convert(tUint64).Uint()
			var buf [10]byte
			l := binary.PutUvarint(buf[:], n)
			return buf[:l]
		},
	}).
	Parse(templateSource))

func Write(w io.Writer, messages []ir.StructRecord) error {
	return tpl.ExecuteTemplate(w, "main", messages)
}
