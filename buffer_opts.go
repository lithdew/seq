package seq

type BufferOption func(b *Buffer)

func WithBufferItemAcked(acked ItemAcked) BufferOption {
	return func(b *Buffer) {
		b.itemAcked = acked
	}
}
