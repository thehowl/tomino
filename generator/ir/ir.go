package ir

import (
	"reflect"
	"strings"
)

type isRecord struct{}

func (isRecord) assertRecord() {}

type ScalarType int

const (
	ScalarTypeInt8 ScalarType = iota
	ScalarTypeInt16
	ScalarTypeInt32
	ScalarTypeInt64

	ScalarTypeUint8
	ScalarTypeUint16
	ScalarTypeUint32
	ScalarTypeUint64

	ScalarTypeFloat32
	ScalarTypeFloat64
	ScalarTypeBool
)

type (
	Record interface {
		assertRecord()
	}

	// u/int/8/16/32/64
	// float32/64
	// bool
	ScalarRecord struct {
		isRecord

		ScalarType
	}

	// structs
	StructRecord struct {
		isRecord
		Fields []StructField
	}

	StructField struct {
		Record Record

		JSONName string
		// TODO: pack these into an integer.
		JSONOmitEmpty  bool
		BinFixed64     bool   // (Binary) Encode as fixed64
		BinFixed32     bool   // (Binary) Encode as fixed32
		BinFieldNum    uint32 // (Binary) max 1<<29-1
		Unsafe         bool   // e.g. if this field is a float.
		WriteEmpty     bool   // write empty structs and lists (default false except for pointers)
		NilElements    bool   // Empty list elements are decoded as nil iff set, otherwise are never nil.
		UseGoogleTypes bool   // If true, decodes Any timestamp and duration to google types.
		CustomMarshal  bool   // whether we have custom MarshalAmino / UnmarshalAmino methods in the original type.
	}

	// slices, arrays
	RepeatedRecord struct {
		Elem Record
	}

	// interfaces
	AnyRecord struct {
		Subset string
	}

	// pointers
	OptionalRecord struct{}

	// []byte, string
	BytesRecord struct{}

	// can be used as a name in [AnyRecord]
	NamedRecord struct {
		Name string
		Elem Record
	}
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
			p.JSONOmitEmpty = true
		}
	}

	// Parse binary tags.
	// NOTE: these get validated later, we don't have TypeInfo yet.
	if binTag == "fixed64" {
		p.BinFixed64 = true
	} else if binTag == "fixed32" {
		p.BinFixed32 = true
	}

	// Parse amino tags.
	aminoTags := strings.Split(aminoTag, ",")
	for _, aminoTag := range aminoTags {
		if aminoTag == "unsafe" {
			p.Unsafe = true
		}
		if aminoTag == "write_empty" {
			p.WriteEmpty = true
		}
		if aminoTag == "nil_elements" {
			p.NilElements = true
		}
	}

	return
}
