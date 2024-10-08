package ir

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type TagFlag byte

const (
	// Encode as fixed64.
	BinFixed64 TagFlag = 1 << iota // `binary:"fixed64"`
	// Encode as fixed32.
	BinFixed32 // `binary:"fixed32"`
	// For encoding floating points.
	Unsafe // `amino:"unsafe"`
	// Write empty structs and lists. (default false except for pointers)
	WriteEmpty // `amino:"write_empty"`
	// Empty list elements are decoded as
	// nil iff set, otherwise are never nil.
	NilElements // `amino:"nil_elements`
	// omitempty JSON field.
	JSONOmitEmpty // `json:",omitempty"`
)

func (t TagFlag) String() string {
	return strings.Join(t.strings(), ",")
}

func (t TagFlag) strings() []string {
	const numTagFlags = 6
	r := make([]string, 0, numTagFlags)
	add := func(flag TagFlag, str string) {
		if t&flag != 0 {
			r = append(r, str)
		}
	}
	add(BinFixed64, "fixed64")
	add(BinFixed32, "fixed32")
	add(Unsafe, "unsafe")
	add(WriteEmpty, "write_empty")
	add(NilElements, "nil_elements")
	add(JSONOmitEmpty, "json_omit_empty")
	return r
}

func (t TagFlag) Has(s string) bool {
	return slices.Contains(t.strings(), s)
}

type (
	Record interface {
		Kind() string
		Validate() error
		assertRecord()
	}

	// u/int/8/16/32/64
	// float32/64
	// bool
	ScalarRecord struct {
		Name string
	}

	// structs
	StructRecord struct {
		Name   string
		Source string
		Fields []StructField
	}

	// individual field of the struct
	StructField struct {
		Name        string
		Record      Record
		TagFlag     TagFlag
		JSONName    string // JSON field name, `json:"<name>"`
		BinFieldNum uint32 // Field number for binary encoding.
	}

	// slices, arrays
	RepeatedRecord struct {
		Elem Record
		// Constrains the size when un/marshalling.
		// If this size is not respected, decoding can be rejected.
		// -1 for slices.
		Size int64
	}

	// interfaces
	AnyRecord struct {
		Subset []string
	}

	// pointers?
	OptionalRecord struct {
		Elem Record
	}

	// []byte, [N]byte, string
	BytesRecord struct {
		// Fixed size of the bytes. If size is not
		// respected when decoding, can be rejected.
		// -1 for non-arrays.
		Size int64
		// A "hint" that this is a string (though nothing should change in marshaling).
		String bool
	}

	// can be used as a name in [AnyRecord]
	NamedRecord struct {
		Name string
		Elem Record
	}
)

func (StructRecord) assertRecord() {}
func (StructRecord) Kind() string  { return "struct" }
func (s StructRecord) Validate() error {
	for _, fld := range s.Fields {
		// can only use fixed flags on appropriate types.
		if fld.TagFlag&BinFixed64 != 0 {
			switch fld.Record {
			case ScalarRecord{Name: "uint64"}, ScalarRecord{Name: "int64"}, ScalarRecord{Name: "float64"}:
			default:
				return fmt.Errorf("invalid record for usage with fixed64: %v", fld.Record)
			}
		}
		if fld.TagFlag&BinFixed32 != 0 {
			switch fld.Record {
			case ScalarRecord{Name: "uint32"}, ScalarRecord{Name: "int32"}, ScalarRecord{Name: "float32"}:
			default:
				return fmt.Errorf("invalid record for usage with fixed32: %v", fld.Record)
			}
		}
		if err := fld.Record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (RepeatedRecord) assertRecord() {}
func (RepeatedRecord) Kind() string  { return "repeated" }
func (rr RepeatedRecord) Validate() error {
	if rr.Size < 0 && rr.Size != -1 {
		return fmt.Errorf("invalid size: %d", rr.Size)
	}
	if rr.Elem == (ScalarRecord{Name: "uint8"}) {
		return errors.New("elem of RepeatedRecord cannot be uint8 (should use BytesRecord instead)")
	}
	return rr.Elem.Validate()
}

func (ScalarRecord) assertRecord() {}
func (ScalarRecord) Kind() string  { return "scalar" }
func (s ScalarRecord) Validate() error {
	switch s.Name {
	case "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return nil
	default:
		return fmt.Errorf("invalid scalar record: %q", s.Name)
	}
}

func (OptionalRecord) assertRecord() {}
func (OptionalRecord) Kind() string  { return "optional" }
func (or OptionalRecord) Validate() error {
	if or.Elem == nil {
		return errors.New("OptionalRecord on nil")
	}
	return or.Elem.Validate()
}

func (BytesRecord) assertRecord() {}
func (BytesRecord) Kind() string  { return "bytes" }
func (br BytesRecord) Validate() error {
	if br.String && br.Size != -1 {
		return errors.New("string BytesRecord must have size = -1")
	}
	if br.Size < 0 && br.Size != -1 {
		return fmt.Errorf("invalid size: %d", br.Size)
	}
	return nil
}

var (
	_ Record = StructRecord{}
	_ Record = ScalarRecord{}
	_ Record = OptionalRecord{}
	_ Record = BytesRecord{}
)

func (p *StructField) ParseTag(tag reflect.StructTag) (skip bool) {
	binTag := tag.Get("binary")
	aminoTag := tag.Get("amino")
	jsonTag := tag.Get("json")

	// If `json:"-"`, don't encode.
	// NOTE: This skips binary as well.
	if jsonTag == "-" {
		return true
	}

	// Get JSON field name.
	jsonTagParts := strings.Split(jsonTag, ",")
	if jsonTagParts[0] != "" {
		p.JSONName = jsonTagParts[0]
	}

	// Get JSON omitempty.
	if len(jsonTagParts) > 1 {
		if jsonTagParts[1] == "omitempty" {
			p.TagFlag |= JSONOmitEmpty
		}
	}

	// Parse binary tags.
	// NOTE: these get validated in Validate().
	if binTag == "fixed64" {
		p.TagFlag |= BinFixed64
	} else if binTag == "fixed32" {
		p.TagFlag |= BinFixed32
	}

	// Parse amino tags.
	aminoTags := strings.Split(aminoTag, ",")
	for _, aminoTag := range aminoTags {
		switch aminoTag {
		case "unsafe":
			p.TagFlag |= Unsafe
		case "write_empty":
			p.TagFlag |= WriteEmpty
		case "nil_elements":
			p.TagFlag |= NilElements
		}
	}

	return
}

func (p StructField) Validate() error {
	switch {
	case p.TagFlag&BinFixed32 != 0 &&
		p.Record != ScalarRecord{"int32"} &&
		p.Record != ScalarRecord{"uint32"}:
		return errors.New("tag fixed32 may only be used on uint32 or int32")
	case p.TagFlag&BinFixed64 != 0 &&
		p.Record != ScalarRecord{"int64"} &&
		p.Record != ScalarRecord{"uint64"}:
		return errors.New("tag fixed64 may only be used on uint64 or int64")
	case p.TagFlag&Unsafe == 0 &&
		(p.Record == ScalarRecord{"float64"} ||
			p.Record == ScalarRecord{"float32"}):
		return errors.New("floating points must be used with the `amino:\"unsafe\"` struct tag")
	}
	return nil
}

func (p StructField) String() string {
	return fmt.Sprintf("%04d=%s[%s] { %v }", p.BinFieldNum, p.Name, p.TagFlag.String(), p.Record)
}

func (p StructField) Tag() []byte {
	const (
		recordTypeVarint = 0
		recordTypeI64    = 1
		recordTypeLen    = 2
		recordTypeI32    = 5
	)

	if _, ok := p.Record.(OptionalRecord); ok {
		panic("StructField.Tag on OptionalRecord; you should handle the OptionalRecord then use .WithRecord on the Elem.")
	}

	// NOTE: here we don't validate whether the type should be a BinFixed64/32.
	// it's done as part of StructField.Validate.
	x := uint64(p.BinFieldNum) << 3
	sr, isScalar := p.Record.(ScalarRecord)
	if p.TagFlag&BinFixed64 != 0 || sr.Name == "float64" {
		x |= recordTypeI64
	} else if p.TagFlag&BinFixed32 != 0 || sr.Name == "float32" {
		x |= recordTypeI32
	} else if isScalar {
		x |= recordTypeVarint
	} else {
		x |= recordTypeLen
	}

	var buf [10]byte
	return buf[:binary.PutUvarint(buf[:], x)]
}

func (p StructField) WithRecord(rec Record) StructField {
	p.Record = rec
	return p
}

func (p StructField) Has(s string) bool {
	return p.TagFlag.Has(s)
}

func (p ScalarRecord) IsUnsigned() bool {
	switch p.Name {
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "int", "int8", "int16", "int32", "int64":
		return false
	default:
		panic(fmt.Sprintf("invalid scalar record for IsUnsigned: %q", p.Name))
	}
}

/* TODO: RepeatedRecord
- If the child is a non-bytelength, then we can encode it in packed form.
- What is ReprType realy? Should we encode []byte as bytes, even when it's a type of bytes?
- Consider what will happen when we have []byte as a result of MarshalAmino (ie returns byte, but it's an array).
- Multidimensional lists (could start off by rejecting them, simply.)

TODO: OptionalRecord
- Support it in StructField.Validate */
