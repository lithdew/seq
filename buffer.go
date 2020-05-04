// Package seq is a fast implementation of sequence buffers described in Go with 100% unit test coverage.
package seq

import (
	"math"
)

// Buffer is a fast, fixed-sized rolling buffer that buffers entries based on an unsigned 16-bit integer.
type Buffer struct {
	oldest  uint16
	next    uint16
	indices []uint32
	entries []interface{}
}

// NewBuffer instantiates a new sequence buffer of size.
func NewBuffer(size uint16) *Buffer {
	if 65536%uint64(size) != 0 {
		panic("BUG: size provided to seq.NewBuffer() must be divisible by 65536")
	}

	return &Buffer{next: 0, oldest: math.MaxUint16, indices: make([]uint32, size), entries: make([]interface{}, size)}
}

// Reset resets the buffer.
func (b *Buffer) Reset() {
	b.next = 0
	b.oldest = math.MaxUint16
	emptyBufferIndices(b.indices)
	emptyBufferEntries(b.entries)
}

// Len returns the size of this buffer.
func (b *Buffer) Len() uint16 {
	return uint16(len(b.entries))
}

// Next returns the next expected sequence number to be inserted/acknowledged by this buffer.
func (b *Buffer) Next() uint16 {
	return b.next
}

// Latest returns the latest sequence number inserted/acknowledged by this buffer.
func (b *Buffer) Latest() uint16 {
	return b.next - 1
}

// Oldest returns the last/oldest consecutively-inserted packet sequence number in this buffer.
func (b *Buffer) Oldest() uint16 {
	return b.oldest
}

// At returns the entry at seq, even if it might be stale.
func (b *Buffer) At(seq uint16) interface{} {
	return b.entries[seq%b.Len()]
}

// Find returns the entry at seq, should seq not be outdated.
func (b *Buffer) Find(seq uint16) interface{} {
	i := seq % b.Len()
	if b.indices[i] == uint32(seq) {
		return b.entries[i]
	}
	return nil
}

// Exists returns whether or not seq is stored in the buffer.
func (b *Buffer) Exists(seq uint16) bool {
	return b.indices[seq%b.Len()] == uint32(seq)
}

// Outdated returns true if seq is capable of being stored in this buffer based on the largest sequence number
// that has been inserted/acknowledged so far.
func (b *Buffer) Outdated(seq uint16) bool {
	return LT(seq, b.next-b.Len())
}

func (b *Buffer) Invalid(seq uint16) bool {
	return GT(seq, b.oldest+b.Len())
}

// Insert inserts a new item into this buffer indexed by seq, should seq not be outdated. It returns true if the
// insertion is successful.
func (b *Buffer) Insert(seq uint16, item interface{}) bool {
	if b.Outdated(seq) {
		return false
	}

	if b.Invalid(seq) {
		return false
	}

	b.updateLatest(seq)

	i := seq % uint16(len(b.entries))
	b.indices[i] = uint32(seq)
	b.entries[i] = item

	b.updateOldest()

	return true
}

func (b *Buffer) updateLatest(seq uint16) {
	if GT(seq+1, b.next) {
		b.RemoveRange(b.next, seq)
		b.next = seq + 1
	}
}

func (b *Buffer) updateOldest() {
	for b.Exists(b.oldest + 1) {
		b.oldest++
	}

	if !b.Exists(b.oldest) {
		b.oldest = b.Latest()
	}
}

// Remove invalidates items and entries stored by the sequence number seq.
func (b *Buffer) Remove(seq uint16) {
	b.indices[seq%b.Len()] = math.MaxUint32
	b.updateOldest()
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

	b.updateOldest()
}

// GenerateLatestBitset32 calls GenerateBitset32 on the latest known sequence number.
func (b *Buffer) GenerateLatestBitset32() (ack uint16, ackBits uint32) {
	ack = b.Latest()
	return ack, b.GenerateBitset32(ack)
}

// GenerateOldestBitset32 calls GenerateBitset32 on the oldest consecutively-known sequence number.
func (b *Buffer) GenerateOldestBitset32() (ack uint16, ackBits uint32) {
	ack = b.Oldest()
	return ack, b.GenerateBitset32(ack)
}

// GenerateBitset32 generates a 32-bit integer representative of a bitset where entries at index i are 1 if there exists
// an entry in the buffer whose sequence number is equal to the largest sequence number inserted/acknowledged so far
// minus i. It returns both the largest sequence number known thus far alongside the bitset as an unsigned 16-bit
// integer and an unsigned 32-bit integer respectively.
func (b *Buffer) GenerateBitset32(ack uint16) (ackBits uint32) {
	for idx, mask := uint16(0), uint32(1); idx < 32; idx, mask = idx+1, mask<<1 {
		seq := ack - idx

		if !b.Exists(seq) {
			continue
		}

		ackBits |= mask
	}

	return ackBits
}
