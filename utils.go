package seq

import (
	"math"
)

var (
	emptyBufferIndexCache [math.MaxUint16]uint32
	emptyBufferEntryCache [math.MaxUint16]interface{}
)

func init() {
	emptyBufferIndexCache[0] = math.MaxUint32
	for i := 1; i < math.MaxUint16; i *= 2 {
		copy(emptyBufferIndexCache[i:], emptyBufferIndexCache[:i])
	}
}

func emptyBufferIndices(indices []uint32) {
	copy(indices[:], emptyBufferIndexCache[:len(indices)])
}

func emptyBufferEntries(entries []interface{}) {
	copy(entries[:], emptyBufferEntryCache[:len(entries)])
}

// Half the max value of an unsigned 16-bit integer.
const HalfMaxUint16 = uint16(math.MaxUint16/2) + 1

// LTE returns whether or not the sequence number a is less than or equal to b. See GT for how this is determined.
func LTE(a, b uint16) bool {
	return a == b || LT(a, b)
}

// GTE returns whether or not the sequence number a is greater than or equal to b. See GT for how this is determined.
func GTE(a, b uint16) bool {
	return a == b || GT(a, b)
}

// LT returns whether or not the sequence number a is less than b. See GT for how this is determined.
func LT(a, b uint16) bool {
	return GT(b, a)
}

// GT returns whether or not the sequence number a is greater than b. This is done by returning whether or not a and b
// are apart by more than 32767 (half the max value of an unsigned 16-bit integer).
func GT(a, b uint16) bool {
	return ((a > b) && (a-b <= HalfMaxUint16)) || ((a < b) && (b-a > HalfMaxUint16))
}
