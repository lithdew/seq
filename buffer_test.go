package seq

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
	"testing/quick"
)

func TestNewBuffer(t *testing.T) {
	require.Panics(t, func() { NewBuffer(102) })

	a, b := NewBuffer(512), NewBuffer(512)

	a.latest, b.latest = 129, 129
	require.Equal(t, a, b)

	a.Reset()
	b.Reset()

	require.Equal(t, a, b)
}

func TestBufferInsertRemove(t *testing.T) {
	buf := NewBuffer(16)

	for i := range rand.Perm(16) {
		require.True(t, buf.Insert(1+uint16(i), true))
		require.EqualValues(t, true, buf.At(1+uint16(i)))
	}

	for i := uint16(1); i <= 16; i++ {
		require.True(t, buf.Exists(i))
		require.EqualValues(t, true, buf.Find(i))
	}

	require.Nil(t, buf.Find(0))
	require.False(t, buf.Exists(0))
	require.False(t, buf.Insert(0, true))

	for i := range rand.Perm(16) {
		buf.Remove(1 + uint16(i))
	}

	for i := uint16(1); i <= 16; i++ {
		require.False(t, buf.Exists(i))
		require.Nil(t, buf.Find(i))
	}
}

func testRemoveRange(t testing.TB) func(uint16, uint8) bool {
	t.Helper()

	return func(start uint16, rawSize uint8) bool {
		size := uint16(rawSize)

		// Make sure the buffer is at least of size 1.

		if size == 0 {
			size++
		}

		// Round size to the nearest power of 2.

		size--
		size |= size >> 1
		size |= size >> 2
		size |= size >> 4
		size |= size >> 8
		size++

		size *= 2

		// Figure out the end of the buffer.

		end := size + start - 1

		buf := NewBuffer(size)
		buf.latest = end + 1

		// Populate buffer items.

		for i := uint16(0); i < size; i++ {
			seq := start + i

			if !assert.True(t, buf.Insert(seq, true)) {
				return false
			}

			if !assert.NotNil(t, buf.Find(seq)) {
				return false
			}

			if !assert.True(t, buf.Exists(seq)) {
				return false
			}
		}

		// Remove range.

		a, b := start+1, end
		buf.RemoveRange(a, b)

		filled := 0
		for _, i := range buf.indices {
			if i != math.MaxUint32 {
				filled++
			}
		}

		if !assert.EqualValues(t, 1, filled) {
			return false
		}

		if !assert.True(t, buf.Exists(start)) {
			return false
		}

		return true
	}
}

func TestBufferRemoveRange(t *testing.T) {
	require.NoError(t, quick.Check(testRemoveRange(t), &quick.Config{MaxCount: 256}))
}

func TestBufferRemoveRangeEdgeCases(t *testing.T) {
	f := testRemoveRange(t)

	for _, size := range []uint8{1, 16, 32, 127, 255} {
		f(0, size)
		f(15673, size)
		f(57152, size)
		f(65535, size)
	}
}

func TestBufferBitset(t *testing.T) {
	buf := NewBuffer(32)

	for i := uint16(0); i < 32; i += 2 {
		buf.Insert(i, true)
	}

	ack, bitset := buf.Bitset()
	require.EqualValues(t, 30, ack)

	for i := uint16(0); i < 32; i, bitset = i+1, bitset>>1 {
		require.True(t, (i&1 != 0 && bitset&1 == 0) || (i&1 == 0 && bitset&1 != 0))
	}

	buf.Insert(buf.latest, true)

	ack, bitset = buf.Bitset()
	require.EqualValues(t, 31, ack)
}

func TestBufferRemoveRangeAll(t *testing.T) {
	start, size := uint16(65535), uint16(512)

	buf := NewBuffer(size)
	buf.latest = size

	for i := uint16(0); i < size; i++ {
		seq := start + i
		buf.indices[seq%size] = uint32(seq)
	}

	buf.RemoveRange(0, size-1)

	for i := uint16(0); i < size; i++ {
		seq := start + i
		require.Nil(t, buf.Find(seq))
	}
}

func BenchmarkTestBufferInsert(b *testing.B) {
	buf := NewBuffer(512)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Insert(uint16(i), 1)
	}
}

func BenchmarkTestBufferRemoveRange(b *testing.B) {
	start, size := uint16(65535), uint16(512)

	buf := NewBuffer(size)
	buf.latest = size

	for i := uint16(0); i < size; i++ {
		seq := start + i
		buf.indices[seq%size] = uint32(seq)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.RemoveRange(start+1, size-2)
	}
}

func BenchmarkTestBufferBitset(b *testing.B) {
	buf := NewBuffer(32)

	for i := uint16(0); i < 32; i += 2 {
		buf.Insert(i, true)
	}

	var (
		ack    uint16
		bitset uint32
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ack, bitset = buf.Bitset()
	}

	_, _ = ack, bitset
}
