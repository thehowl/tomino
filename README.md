# tomino

A code-generated drop-in replacement for [amino], creating minimal 0-dependency
marshalers with no reflection. **See the [status](#status).**

```mermaid
flowchart TD
    GoSource[Go source files] --> Go
    Go["Parsed source
(_x/tool/packages_)"] -->|generator| IR["Intermediate Representation
(_generator/ir_)"]
    IR -->|generator/target/go| OutGo[Go Output]
    IR -->|TODO| OutJS[JavaScript Output]
    IR -->|TODO| OutZig[Zig Output]
    IR --> Etc[...]
```

## Objectives

- `< 5000` lines of non-test code, for go-targeted generation.
- Simple IR to write Amino encoders and decoders in other languages, both
  directly in Go (similarly to [generator/targets/go](./generator/targets/go),
  or by marshaling into Amino/JSON and parsing in another language.
- No reflection needed; support for interfaces using type switches on all
  possible registered types.
- Inspectable and simple generated code (one target file, no dependencies).
- Faster than Amino+Reflect and Amino+Protobuf. (And maybe, even protobuf.)
- Eventually, hand-written specifications on binary encoding, go type parsing,
  and IR (and how to build support in other languages).

We can make it happen.

## Status

This is still in very early stages, but it has some promising results (oct 5, 2024):

```
goos: linux
goarch: amd64
pkg: github.com/thehowl/tomino/tests
cpu: AMD Ryzen 7 7840U w/ Radeon  780M Graphics
BenchmarkMarshalers/tomino/bytes_1_000         	 7047403	       144.5 ns/op	    1088 B/op	       2 allocs/op
BenchmarkMarshalers/tomino/bytes_1_000_000     	   14998	     71262 ns/op	 1007680 B/op	       2 allocs/op
BenchmarkMarshalers/tomino/empty               	43339066	        27.72 ns/op	      64 B/op	       1 allocs/op
BenchmarkMarshalers/tomino/fixed               	42246316	        28.32 ns/op	      64 B/op	       1 allocs/op
BenchmarkMarshalers/tomino/ptr_-1337           	40327893	        29.70 ns/op	      64 B/op	       1 allocs/op
BenchmarkMarshalers/tomino/slice               	24488661	        43.95 ns/op	      64 B/op	       1 allocs/op
BenchmarkMarshalers/tomino/time_duration       	24913022	        41.78 ns/op	      64 B/op	       1 allocs/op
BenchmarkMarshalers/tomino_prealloc/bytes_1_000         	69462626	        17.07 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshalers/tomino_prealloc/bytes_1_000_000     	   15192	     71695 ns/op	 1007616 B/op	       1 allocs/op
BenchmarkMarshalers/tomino_prealloc/empty               	195221311	         6.091 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshalers/tomino_prealloc/fixed               	176251285	         6.773 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshalers/tomino_prealloc/ptr_-1337           	147158512	         8.137 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshalers/tomino_prealloc/slice               	47009396	        21.80 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshalers/tomino_prealloc/time_duration       	59409352	        20.14 ns/op	       0 B/op	       0 allocs/op
BenchmarkMarshalers/amino/bytes_1_000                   	 1319538	       891.0 ns/op	    2512 B/op	      16 allocs/op
BenchmarkMarshalers/amino/bytes_1_000_000               	    8067	    143878 ns/op	 2015696 B/op	      16 allocs/op
BenchmarkMarshalers/amino/empty                         	 2431149	       495.1 ns/op	     432 B/op	      12 allocs/op
BenchmarkMarshalers/amino/fixed                         	 2108006	       563.4 ns/op	     528 B/op	      15 allocs/op
BenchmarkMarshalers/amino/ptr_-1337                     	 2021954	       575.1 ns/op	     528 B/op	      15 allocs/op
BenchmarkMarshalers/amino/slice                         	 1005511	      1182 ns/op	    1168 B/op	      35 allocs/op
BenchmarkMarshalers/amino/time_duration                 	 1615058	       739.3 ns/op	     752 B/op	      21 allocs/op
PASS
ok  	github.com/thehowl/tomino/tests	38.245s
```

You can find the benchmarks in [tests/amino_test.go](./tests/amino_test.go); they
are broken down into the following "marshalers":

- `tomino` is a simple call to `MarshalBinary`, which always does an allocation to
  a 64-byte buffer.
- `tomino_prealloc` is a call to `AppendBinary`, re-using the same buffer. It shows
  the raw encoding power, assuming a best-case scenario where we can re-use a
  buffer.
- `amino` calls the same structs as before, but using `amino.Marshal`, as a
  baseline benchmark for comparison.

These are in no way definitive results (I have been writing performant code but
not inspecting to find any specific slowness; similarly for amino, it may be
better to at least test it out with proto3 bindings). But as you can see, the
results are at the very least promising: significantly less heap allocations,
and performance improvements varying from 2x all the way up to ~100x.

> The performance improvements against amino can vary greatly, mostly as a
> consequence of inevitable allocations. For instance, the `bytes_1_000_000`
> benchmark is close, because for both most of the time will simply be spent
> allocating the 1-MB slice, anyway.
>
> On the other hand, where encoder/decoder logic is heavily involved,
> (and consequently, `B/op` for both is lower), tomino generally shines, with
> speed improvements mostly sitting around ~20x.

There still isn't a decoder. I want to try to get all the current feature-set of
amino ported over to code generating go marshalers, before starting on the
decoder, and then heading onto other languages.

If the vision with this thing succeeds, we should at the very least have a tool
to create amino marshalers/unmarshalers outside of Go. If this vision succeeds
tremendously, we might be able to replace the current usages in
[gnolang/gno](https://github.com/gnolang/gno) with this package.

## Binary encoding

Heavily drawn from [protobuf's documentation "Encoding" page](https://protobuf.dev/programming-guides/encoding/#cheat-sheet).

This section only tackles how the binary encoding of the format works. The
[Language specification][language] specifies instead how the Go specification
can be parsed to create decoders, depending on the types.

XXX: additional constraints. what does amino constrain that protobuf doesn't? i
imagine we want to make sure to make the result deterministic and to say that
ambiguities shouldn't happen; like out-of-order repeated fields, or
out-of-order fields.

### Base 128 varint

TODO: copy over section from protobuf page.

### Message structure

A message is a series of tag-value pairs. Each pair is called a "record".
The tag determines the "field number" and the "type" of the value; which in
turn determines its length.

ID | Name    | Value length
---|---------|-------------------------------------
0  | varint  | Variable (see [varint])
1  | i64     | 8 bytes
2  | len     | [varint] length N + N bytes
5  | i32     | 4 bytes

The field number and the type ID are encoded as a varint, packed together via
the formula `(field_number << 3) | type_id`.

Multiple messages with the same field_number may appear. These generally
indicate a repeated field, like a slice or array.

### Uses for scalar-type values

- varint: encodes all kind of signed and unsigned integer values, including
    booleans. "lower" unsigned values use fewer bytes.
- i64: encoding fixed-size int64 values, including float32's.
- i32: encoding fixed-size int32 values, including float64's.

### Uses for len-type values

The decoder will determine, based on the underlying language type, how to parse
the bytes. Here are common usages:

- raw bytes: useful for strings, and byte sequences.
- submessages: encodings of other messages.
- packed messages: messages of repeated values of type varint, i64 or i32, can
    have their values concatenated together and placed in a len-type value. this
    is called a "packed" message.

## Language specification

The Language specification allows to parse source code, so that it can be used
to create tomino encoders and decoders. It is the tomino's parallel to the
[protocol buffers language](https://protobuf.dev/reference/protobuf/edition-2023-spec/).

This sounds all well and good. The caveat is that tomino requires a compliant
parser of the [Go language specification](https://go.dev/ref/spec). tomino only
handles a subset; and a full compiler of the Go programming language is not
required. But your parser needs not only to be able to understand type
declarations, but also resolve references to external named types.

(Note: this may change if the idea of tomino gains traction and someone is
interested in developing a Go-compatible grammar and specification of the
importer.)

Assuming you have a working Go parser and type-checker for resolving type names,
let's get started.

An encoder and decoder may be generated for any
[type declaration](https://go.dev/ref/spec#Type_declarations).
Names are resolved, until we arrive at a

[amino]: https://github.com/gnolang/gno/tree/master/tm2/pkg/amino
[language]: #language-specification
[varint]: #base-128-varint
