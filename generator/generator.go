package generator

import (
	"fmt"
	"go/types"
	"reflect"

	"github.com/thehowl/tomino/generator/ir"
)

// Parse XXX
func Parse(obj types.Object) (ir.StructRecord, error) {
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return ir.StructRecord{}, fmt.Errorf("invalid symbol: %T", obj)
	}

	// TODO: does this work with aliases? (maybe it shouldn't.)
	tp := tn.Type()
	rec, err := parse(tp)
	if err != nil {
		return ir.StructRecord{}, err
	}
	return rec.(ir.StructRecord), nil
}

func parse(tp types.Type) (ir.Record, error) {
	// TODO: change to custom error type.
	switch tp := tp.(type) {
	case *types.Struct:
		flds := make([]ir.StructField, 0, tp.NumFields())
		for i := 0; i < tp.NumFields(); i++ {
			fld := tp.Field(i)
			if !fld.Exported() {
				continue
			}
			tag := tp.Tag(i)
			sf := ir.StructField{
				Name:     fld.Name(),
				JSONName: fld.Name(),
				// TODO: would be best if this weren't so brittle.
				// (it's how amino does it.)
				BinFieldNum: uint32(len(flds) + 1),
			}
			skip := sf.ParseTag(reflect.StructTag(tag))
			if skip {
				continue
			}
			var err error
			sf.Record, err = parse(fld.Type())
			if err != nil {
				return nil, err
			}
			flds = append(flds, sf)
		}
		return ir.StructRecord{Fields: flds}, nil
	case *types.Basic:
		// TODO: does this understand rune == int32 and byte == uint8?
		// if not, let's use tp.Kind() instead.
		sr := ir.ScalarRecord{Name: tp.Name()}
		switch sr.Name {
		case "byte":
			sr.Name = "uint8"
		case "rune":
			sr.Name = "int32"
		case "string":
			return ir.BytesRecord{String: true, Size: -1}, nil
		}

		return sr, sr.Validate()
	case *types.Array:
		if isUint8(tp.Elem()) {
			return ir.BytesRecord{Size: tp.Len()}, nil
		}
		panic("not implemented")
	case *types.Slice:
		if isUint8(tp.Elem()) {
			return ir.BytesRecord{Size: -1}, nil
		}
		panic("not implemented")
	case *types.Interface:
		panic("not implemented")
	case *types.Pointer:
		if _, isPtr := tp.Elem().Underlying().(*types.Pointer); isPtr {
			return nil, fmt.Errorf("type %v is pointer of pointer", tp.String())
		}
		v, err := parse(tp.Elem())
		if err != nil {
			return nil, err
		}
		return ir.OptionalRecord{Elem: v}, nil
	case *types.Named:
		if obj := tp.Obj(); obj.Pkg().Path() == "time" {
			timeFields := []ir.StructField{
				{
					Name:        "Seconds",
					Record:      ir.ScalarRecord{Name: "uint64"},
					JSONName:    "seconds",
					BinFieldNum: 1,
				},
				{
					Name:        "Nanoseconds",
					Record:      ir.ScalarRecord{Name: "uint32"},
					JSONName:    "nanoseconds",
					BinFieldNum: 2,
				},
			}
			// Encode Time and Duration differently than we would do otherwise.
			// The Go converter will handle the details of converting from the orig
			// type.
			// Seconds and Nanoseconds are encoded as uints to encode them using
			// Uvarint; but they are signed.
			switch tp.Obj().Name() {
			case "Duration":
				return ir.StructRecord{
					Name:   "Duration",
					Source: "time.Duration",
					Fields: timeFields,
				}, nil
			case "Time":
				return ir.StructRecord{
					Name:   "Time",
					Source: "time.Time",
					Fields: timeFields,
				}, nil
			}
		}

		// TODO: should centralize names in a registry so we re-use encoders.
		// TODO: should understand a type having AminoMarshal / AminoUnmarshal.
		parsed, err := parse(tp.Underlying())
		if err != nil {
			return nil, err
		}
		if str, ok := parsed.(ir.StructRecord); ok {
			str.Name = tp.Obj().Name()
			str.Source = tp.String()
			return str, nil
		}
		return parsed, err
	default:
		return nil, fmt.Errorf("unsupported type: %T (%v)", tp, tp)
	}
}

func isUint8(tp types.Type) bool {
	if bas, ok := tp.(*types.Basic); ok {
		return bas.Kind() == types.Byte
	}
	return false
}
