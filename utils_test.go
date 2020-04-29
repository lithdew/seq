package seq

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
	"testing/quick"
)

func TestUint16Rollover(t *testing.T) {
	a := uint16(0)

	a--
	require.EqualValues(t, math.MaxUint16, a)

	a++
	require.EqualValues(t, 0, a)
}

func randBufferEntries(indices []uint32) {
	for i := range indices {
		indices[i] = rand.Uint32()
	}
}

func TestEmptyBufferEntries(t *testing.T) {
	entries := make([]uint32, math.MaxUint16)

	f := func(count uint16) bool {
		randBufferEntries(entries[:count])
		emptyBufferIndices(entries[:count])
		return assert.EqualValues(t, emptyBufferIndexCache[:count], entries[:count])
	}

	require.NoError(t, quick.Check(f, nil))
}
