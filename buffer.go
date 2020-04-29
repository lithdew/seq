// Package seq is a fast implementation of sequence buffers described in Go with 100% unit test coverage.
package seq

import "math"

// Buffer is a fast, fixed-sized rolling buffer that buffers entries based on an unsigned 16-bit integer.
type Buffer struct {
	latest  uint16
	indices []uint32
	entries []interface{}
}

// NewBuffer instantiates a new sequence buffer of size.
func NewBuffer(size uint16) *Buffer {
	if size&(size-1) != 0 {
		panic("BUG: size provided to seq.NewBuffer() must be a power of two")
	}

	return &Buffer{latest: 0, indices: make([]uint32, size), entries: make([]interface{}, size)}
}

// Reset resets the buffer.
func (b *Buffer) Reset() {
	b.latest = 0
	emptyBufferIndices(b.indices)
	emptyBufferEntries(b.entries)
}

// At returns the entry at seq, even if it might be stale.
func (b *Buffer) At(seq uint16) interface{} {
	return b.entries[seq%uint16(len(b.entries))]
}

// Find returns the entry at seq, should seq not be outdated.
func (b *Buffer) Find(seq uint16) interface{} {
	i := seq % uint16(len(b.entries))
	if b.indices[i] == uint32(seq) {
		return b.entries[i]
	}
	return nil
}

// Exists returns whether or not seq is stored in the buffer.
func (b *Buffer) Exists(seq uint16) bool {
	return b.indices[seq%uint16(len(b.entries))] == uint32(seq)
}

// Outdated returns true if seq is capable of being stored in this buffer based on the largest sequence number
// that has been inserted/acknowledged so far.
func (b *Buffer) Outdated(seq uint16) bool {
	return LT(seq, b.latest-uint16(len(b.entries)))
}

// Insert inserts a new item into this buffer indexed by seq, should seq not be outdated. It returns true if the
// insertion is successful.
func (b *Buffer) Insert(seq uint16, item interface{}) bool {
	if b.Outdated(seq) {
		return false
	}

	if GT(seq+1, b.latest) {
		b.RemoveRange(b.latest, seq)
		b.latest = seq + 1
	}

	i := seq % uint16(len(b.entries))
	b.indices[i] = uint32(seq)
	b.entries[i] = item
	return true
}

// Remove invalidates items and entries stored by the sequence number seq.
func (b *Buffer) Remove(seq uint16) {
	b.indices[seq%uint16(len(b.entries))] = math.MaxUint32
}

// RemoveRange invalidates all items and entries with sequence numbers in the range [start, end].
func (b *Buffer) RemoveRange(start, end uint16) {
	count, size := end-start+1, uint16(len(b.entries))

	if count >= size {
		emptyBufferIndices(b.indices)
		return
	}

	first := b.indices[start%size:]
	length := uint16(len(first))

	if count <= length {
		emptyBufferIndices(first[:count])
		return
	}

	second := b.indices[:count-length]

	emptyBufferIndices(first)
	emptyBufferIndices(second)
}

// Bitset generates a 32-bit integer representative of a bitset where entries at index i are 1 if there exists an entry
// in the buffer whose sequence number is equal to the largest sequence number inserted/acknowledged so far minus i. It
// returns both the largest sequence number known thus far alongside the bitset as an unsigned 16-bit integer and a
// unsigned 32-bit integer respectively.
func (b *Buffer) Bitset() (ack uint16, ackBits uint32) {
	ack = b.latest - 1

	for idx, mask := uint16(0), uint32(1); idx < 32; idx, mask = idx+1, mask<<1 {
		seq := ack - idx

		if !b.Exists(seq) {
			continue
		}

		ackBits |= mask
	}

	return ack, ackBits
}
