package gotarget

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/thehowl/tomino/generator/ir"

	_ "embed"
)

//go:embed template.tmpl
var templateSource string

var tUint64 = reflect.TypeOf(uint64(0))

const hextable = "0123456789abcdef"

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
		"gotag": func(b []uint8) string {
			if len(b) == 1 {
				// one-byte tag: since go evaluates constant expressions at compile time,
				// break this down into the components.
				n := b[0]
				return fmt.Sprintf("(%d << 3) | %d /* 0x%02x */", n>>3, n&(1<<3-1), n)
			}

			// multiple bytes: write them out as "0xde, 0xad"
			var bld strings.Builder
			bld.Grow(len(b)*6 - 2)
			for pos, n := range b {
				bld.WriteString("0x")
				bld.WriteByte(hextable[n>>4])
				bld.WriteByte(hextable[n&15])
				if pos != len(b)-1 {
					bld.WriteString(", ")
				}
			}
			return bld.String()
		},
	}).
	Parse(templateSource))

func Write(w io.Writer, messages []ir.StructRecord) error {
	return tpl.ExecuteTemplate(w, "main", messages)
}
