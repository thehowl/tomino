package ir

import (
	"encoding/binary"
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
		assertRecord()
	}

	// u/int/8/16/32/64
	// float32/64
	// bool
	ScalarRecord struct {
		// TODO: custom type/validation.
		Name string
	}

	// structs
	StructRecord struct {
		Name   string
		Source string
		Fields []StructField
	}

	StructField struct {
		Name   string
		Record Record
		TagFlag
		JSONName    string // JSON field name, `json:"<name>"`
		BinFieldNum uint32 // Field number for binary encoding.
	}

	// slices, arrays
	RepeatedRecord struct {
		Elem Record
	}

	// interfaces
	AnyRecord struct {
		Subset []string
	}

	// pointers?
	OptionalRecord struct {
		Elem Record
	}

	// []byte, string
	BytesRecord struct {
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

func (ScalarRecord) assertRecord() {}
func (ScalarRecord) Kind() string  { return "scalar" }

func (OptionalRecord) assertRecord() {}
func (OptionalRecord) Kind() string  { return "optional" }

func (BytesRecord) assertRecord() {}
func (BytesRecord) Kind() string  { return "bytes" }

var (
	_ Record = StructRecord{}
	_ Record = ScalarRecord{}
	_ Record = OptionalRecord{}
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
	// NOTE: these get validated later, we don't have TypeInfo yet.
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

func (p *StructField) String() string {
	return fmt.Sprintf("%04d=%s[%s] { %v }", p.BinFieldNum, p.Name, p.TagFlag.String(), p.Record)
}

func (p *StructField) Tag() []byte {
	const (
		recordTypeVarint = 0
		recordTypeI64    = 1
		recordTypeLen    = 2
		recordTypeI32    = 5
	)

	// NOTE: here we don't validate whether the type should be a BinFixed64/32
	// (TODO)
	x := uint64(p.BinFieldNum) << 3
	sr, _ := p.Record.(*ScalarRecord)
	if p.TagFlag&BinFixed64 != 0 || (sr != nil && sr.Name == "float64") {
		x |= recordTypeI64
	} else if p.TagFlag&BinFixed32 != 0 || (sr != nil && sr.Name == "float32") {
		x |= recordTypeI32
	} else if _, ok := p.Record.(*ScalarRecord); ok {
		x |= recordTypeVarint
	} else {
		x |= recordTypeLen
	}

	var buf [10]byte
	return buf[:binary.PutUvarint(buf[:], x)]
}

func (p *ScalarRecord) IsUnsigned() bool {
	switch p.Name {
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "int", "int8", "int16", "int32", "int64":
		return false
	default:
		panic(fmt.Sprintf("invalid scalar record for IsUnsigned: %q", p.Name))
	}
}
