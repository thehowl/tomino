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
		switch {
		case sr.Name == "string":
			return ir.BytesRecord{String: true}, nil
		case sr.Validate() == nil:
			return sr, nil
		default:
			return nil, fmt.Errorf("unsupported basic type: %v", tp.Name())
		}
	case *types.Array:
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
	case *types.Slice:
		panic("not implemented")
	case *types.Named:
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
