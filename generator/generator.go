package generator

import (
	"fmt"
	"go/types"
	"reflect"
	"slices"

	"github.com/thehowl/tomino/generator/ir"
)

// ParseTypes construct ir.Records based on the passed types.
// The resulting ir.Records can be used to create code-generators in various
// programming languages.
//
// The resulting top-level ir.Records are only StructRecords or AnyRecords.
func ParseTypes(symbols []types.Object) ([]ir.Record, error) {
	pctx := &parseCtx{
		syms: symbols,
		recs: make([]ir.Record, len(symbols)),
	}
	for i := 0; i < len(pctx.syms); i++ {
		sym := pctx.syms[i]
		tn, ok := sym.(*types.TypeName)
		if !ok {
			return nil, fmt.Errorf("invalid symbol type passed to ParseTypes: %T", sym)
		}

		// TODO: does this work with aliases? (maybe it shouldn't.)
		tp := tn.Type()
		rec, err := pctx.parse(tp, true)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", tn.Id(), err)
		}
		pctx.recs[i] = rec
	}
	return pctx.recs, nil
}

type parseCtx struct {
	// len(syms) == len(recs)
	syms []types.Object
	recs []ir.Record
}

func (ctx *parseCtx) parse(tp types.Type, isRoot bool) (ir.Record, error) {
	// TODO: change to custom error type.
	switch tp := tp.(type) {
	case *types.Basic:
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
	case *types.Pointer:
		if _, isPtr := tp.Elem().Underlying().(*types.Pointer); isPtr {
			return nil, fmt.Errorf("type %v is pointer of pointer", tp.String())
		}
		v, err := ctx.parse(tp.Elem(), false)
		if err != nil {
			return nil, err
		}
		return ir.OptionalRecord{Elem: v}, nil
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
			sf.Record, err = ctx.parse(fld.Type(), false)
			if err != nil {
				return nil, err
			}
			flds = append(flds, sf)
		}
		return ir.StructRecord{Fields: flds}, nil
	case *types.Array:
		if isUint8(tp.Elem()) {
			return ir.BytesRecord{Size: tp.Len()}, nil
		}
		elem, err := ctx.parse(tp.Elem(), false)
		if err != nil {
			return nil, err
		}
		return ir.RepeatedRecord{Elem: elem, Size: tp.Len()}, nil
	case *types.Slice:
		if isUint8(tp.Elem()) {
			return ir.BytesRecord{Size: -1}, nil
		}
		elem, err := ctx.parse(tp.Elem(), false)
		if err != nil {
			return nil, err
		}
		return ir.RepeatedRecord{Elem: elem, Size: -1}, nil
	case *types.Interface:
		if tp.NumMethods() != 0 {
			return nil, fmt.Errorf("only the empty interface is allowed as a non-named interface")
		}
		return ir.AnyRecord{}, nil
	case *types.Named:
		if sr, ok := findWellKnown(tp); ok {
			return sr, nil
		}

		// TODO: should understand a type having AminoMarshal / AminoUnmarshal.

		if !isRoot {
			// If this is not a root parse(), and we have a type which we're
			// already going to parse, use that instead.
			target := slices.IndexFunc(ctx.syms, func(obj2 types.Object) bool {
				if obj2.Id() == "" {
					panic("empty obj id, should not happen")
				}
				return obj2.Id() == tp.Obj().Id()
			})
			// lazily add interfaces to the types to process.
			if _, ok := tp.Underlying().(*types.Interface); target < 0 && ok {
				return nil, fmt.Errorf(
					"interface %v is not listed as a symbol to parse; interfaces need to be explicitly required",
					tp.String())
			}
			if target >= 0 {
				return ir.NamedRecord{PtrElem: &ctx.recs[target]}, nil
			}
		} else {
			// interface at root - create the subset.
			if iface, ok := tp.Underlying().(*types.Interface); ok {
				irec := ir.AnyRecord{
					Name:   tp.Obj().Name(),
					Source: tp.String(),
				}
				for idx, other := range ctx.syms {
					if _, ok := other.Type().Underlying().(*types.Struct); !ok {
						// Only work on concrete types.
						continue
					}
					if types.Implements(other.Type(), iface) {
						irec.Subset = append(irec.Subset, ir.NamedRecord{
							PtrElem: &ctx.recs[idx],
						})
					}
				}
				return irec, nil
			}
		}

		parsed, err := ctx.parse(tp.Underlying(), false)
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

// isUint8 determines whether the given type is a uint8 or alias (like byte).
func isUint8(tp types.Type) bool {
	if bas, ok := tp.(*types.Basic); ok {
		return bas.Kind() == types.Byte
	}
	return false
}

// findWellKnown handles the case of amino's "well-known" types; ie. the
// time.Time and time.Duration types. If those are encountered, the underlying
// StructRecord to be generated (and consequently, the encoding) will be
// different.
func findWellKnown(tp *types.Named) (ir.StructRecord, bool) {
	obj := tp.Obj()
	if obj.Pkg().Path() != "time" {
		return ir.StructRecord{}, false
	}
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
		}, true
	case "Time":
		return ir.StructRecord{
			Name:   "Time",
			Source: "time.Time",
			Fields: timeFields,
		}, true
	default:
		return ir.StructRecord{}, false
	}
}
