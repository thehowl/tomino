package tests

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tomtypes "github.com/thehowl/tomino/tests/golden"
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestMarshalerCompatibility(t *testing.T) {
	// deterministic, good random source
	rnd := rand.New(rand.NewChaCha8(
		sha256.Sum256([]byte("the quick brown fox jumps over the lazy dog.")),
	))

	tm := map[string]tomtypes.TestTypeMessage{
		"empty":     {},
		"ptr_0":     {IntPtr: ptrTo(0)},
		"ptr_123":   {IntPtr: ptrTo(123)},
		"ptr_-1337": {IntPtr: ptrTo(-1337)},
		"time_duration": func() tomtypes.TestTypeMessage {
			v := tomtypes.TestTypeMessage{}
			v.Duration.Nanoseconds = 1
			p := -1337
			v.Duration.Seconds = uint64(p)
			v.Time.Seconds = 900000
			return v
		}(),
		"bytes":           {Bytes: []byte{1, 2, 3, 4}},
		"bytes_1_000":     {Bytes: randBytes(rnd, 1000)},
		"bytes_1_000_000": {Bytes: randBytes(rnd, 1_000_000)},
		"bytes_arr":       {ByteArr: &[4]byte{1, 2, 3, 4}},
		"slice": {Slice: []struct {
			A int `json:"A"`
			B int `json:"B"`
		}{{1, 5}, {0, 4}, {1337, 0}}},
		"fixed": {FixedUint: 0xdeadbeef},
	}

	for _, name := range sortedMapKeys(tm) {
		v := tm[name]
		t.Run(name, func(t *testing.T) {
			aminoRes, err := amino.Marshal(v)
			require.NoError(t, err)

			tominoRes, err := v.MarshalBinary()
			require.NoError(t, err)

			if name == "slice" {
				t.Log("\n" + hex.Dump(tominoRes))
			}

			if len(aminoRes) == 0 && len(tominoRes) == 0 {
				// return here, to avoid any incosistencies like amino returning nil
				// and tomino returning []byte{}.
				return
			}

			assert.Equal(t, aminoRes, tominoRes)
		})
	}
}

func randBytes(r *rand.Rand, n int) []byte {
	b := make([]byte, n)
	p := r.Uint64()
	for i := 0; i < n; i++ {
		b[i] = byte(p)
		p >>= 4
		if p == 0 {
			p = r.Uint64()
		}
	}
	return b
}

func sortedMapKeys[K cmp.Ordered, V any, M ~map[K]V](m M) (k []K) {
	k = make([]K, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	slices.Sort(k)
	return k
}

func BenchmarkMarshalers(b *testing.B) {
	// deterministic, good random source
	rnd := rand.New(rand.NewChaCha8(
		sha256.Sum256([]byte("the quick brown fox jumps over the lazy dog.")),
	))

	tm := map[string]tomtypes.TestTypeMessage{
		"empty":     {},
		"ptr_-1337": {IntPtr: ptrTo(-1337)},
		"time_duration": func() tomtypes.TestTypeMessage {
			v := tomtypes.TestTypeMessage{}
			v.Duration.Nanoseconds = 1
			p := -1337
			v.Duration.Seconds = uint64(p)
			v.Time.Seconds = 900000
			return v
		}(),
		"bytes_1_000":     {Bytes: randBytes(rnd, 1000)},
		"bytes_1_000_000": {Bytes: randBytes(rnd, 1_000_000)},
		"slice": {Slice: []struct {
			A int `json:"A"`
			B int `json:"B"`
		}{{1, 5}, {0, 4}, {1337, 0}}},
		"fixed": {FixedUint: 0xdeadbeef},
	}

	b.Run("tomino", func(b *testing.B) {
		for _, name := range sortedMapKeys(tm) {
			v := tm[name]
			var dst []byte
			_ = dst
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					// assign to dst to prevent optimizations
					dst, _ = v.MarshalBinary()
				}
			})
		}
	})
	buf := make([]byte, 16<<10)
	b.Run("tomino_prealloc", func(b *testing.B) {
		for _, name := range sortedMapKeys(tm) {
			v := tm[name]
			var dst []byte
			_ = dst
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					// assign to dst to prevent optimizations
					dst, _ = v.AppendBinary(buf[:0])
				}
			})
		}
	})
	b.Run("amino", func(b *testing.B) {
		for _, name := range sortedMapKeys(tm) {
			v := tm[name]
			var dst []byte
			_ = dst
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					// assign to dst to prevent optimizations
					dst, _ = amino.Marshal(v)
				}
			})
		}
	})
}
