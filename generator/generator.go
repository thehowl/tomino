package generator

import (
	"fmt"
	"go/types"
	"reflect"

	"github.com/thehowl/tomino/generator/ir"
)

// Parse XXX
func Parse(obj types.Object) (*ir.StructRecord, error) {
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("invalid symbol: %T", obj)
	}

	// TODO: does this work with aliases? (maybe it shouldn't.)
	tp := tn.Type().Underlying()
	rec, err := parse(tp)
	if err != nil {
		return nil, err
	}
	return rec.(*ir.StructRecord), nil
}

func parse(tp types.Type) (ir.Record, error) {
	switch tp := tp.(type) {
	case *types.Struct:
		flds := make([]ir.StructField, 0, tp.NumFields())
		for i := 0; i <= tp.NumFields(); i++ {
			fld := tp.Field(i)
			if !fld.Exported() {
				continue
			}
			tag := tp.Tag(i)
			sf := ir.StructField{
				JSONName: fld.Name(),
				// would be best if this weren't so brittle.
				BinFieldNum: uint32(len(flds) + 1),
			}
			sf.ParseTag(reflect.StructTag(tag))
			flds = append(flds, sf)
		}
		sr := &ir.StructRecord{}
	case *types.Basic:
		panic("not implemented")
	case *types.Array:
		panic("not implemented")
	case *types.Interface:
		panic("not implemented")
	case *types.Pointer:
		panic("not implemented")
	case *types.Slice:
		panic("not implemented")
	default:
		return nil, fmt.Errorf("unsupported object: %v", obj)
	}

	return nil, nil
}
