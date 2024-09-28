package tests

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tomtypes "github.com/thehowl/tomino/tests/golden"
)

func TestMarshalerCompatibility(t *testing.T) {
	v := tomtypes.TestTypeMessage{}
	v.Duration.Nanoseconds = 1
	p := -1337
	v.Duration.Seconds = uint64(p)
	v.Time.Seconds = 900000

	aminoRes, err := amino.Marshal(v)
	require.NoError(t, err)

	tominoRes, err := v.MarshalBinary()
	require.NoError(t, err)

	assert.Equal(t, aminoRes, tominoRes)
}
