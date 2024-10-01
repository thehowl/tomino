package tests

import (
	"cmp"
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
	tm := map[string]tomtypes.TestTypeMessage{
		"empty":   {},
		"ptr_0":   {IntPtr: ptrTo(0)},
		"ptr_123": {IntPtr: ptrTo(123)},
		"time_duration": func() tomtypes.TestTypeMessage {
			v := tomtypes.TestTypeMessage{}
			v.Duration.Nanoseconds = 1
			p := -1337
			v.Duration.Seconds = uint64(p)
			v.Time.Seconds = 900000
			return v
		}(),
	}

	for _, name := range sortedMapKeys(tm) {
		v := tm[name]
		t.Run(name, func(t *testing.T) {
			aminoRes, err := amino.Marshal(v)
			require.NoError(t, err)

			tominoRes, err := v.MarshalBinary()
			require.NoError(t, err)

			assert.Equal(t, aminoRes, tominoRes)
		})
	}
}

func sortedMapKeys[K cmp.Ordered, V any, M ~map[K]V](m M) (k []K) {
	k = make([]K, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	slices.Sort(k)
	return k
}
