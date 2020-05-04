# seq

[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](LICENSE)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lithdew/seq)
[![Discord Chat](https://img.shields.io/discord/697002823123992617)](https://discord.gg/HZEbkeQ)

A fast implementation of sequence buffers described in this [blog post by Glenn Fiedler](https://gafferongames.com/post/reliable_ordered_messages/) in Go with 100% unit test coverage.

This was built for the purpose of creating reliable UDP networking protocols, where sequence buffers may be used as an efficient, resilient, fixed-sized rolling buffer for:

1. tracking metadata over sent/received packets,
2. ordering packets received from an unreliable stream, or
3. tracking acknowledgement over the recipient of sent packets from peers.

The sequence numbers used to buffer entries are fixed to be unsigned 16-bit integers, as larger amounts of entries are redundant and would provide a negligible improvement to your packet acknowledgement system.

## Notes

The size of the buffer must be divisible by the max value of an unsigned 16-bit integer (65536), otherwise data buffered by sequence numbers would not wrap around the entire buffer. This was encountered while writing tests for this library.

The method `RemoveRange` was benchmarked and optimized over the sequence buffer implementation in the reference C codebase [reliable.io](https://github.com/networkprotocol/reliable.io) to use a few `memcpy` calls over for loops.

The actual sequences and buffered data are stored in two separate, contiguous slices so that entries that have popped from the rolling buffer will remain as stale memory that may optionally be garbage-collected later.

## Setup

```
go get github.com/lithdew/seq
```

## Benchmarks

```
$ go test -bench=. -benchtime=10s
goos: linux
goarch: amd64
pkg: github.com/lithdew/seq
BenchmarkTestBufferInsert-8             499233990               24.5 ns/op             0 B/op          0 allocs/op
BenchmarkTestBufferRemoveRange-8        233703596               52.3 ns/op             0 B/op          0 allocs/op
BenchmarkTestBufferGenerateBitset32-8   74007890                142 ns/op              0 B/op          0 allocs/op
```