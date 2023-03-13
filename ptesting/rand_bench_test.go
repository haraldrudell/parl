/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptesting

import (
	cryptorand "crypto/rand"
	"math/rand"
	"testing"
	"unsafe"
)

// go test -run=^# -bench=BenchmarkRand ./ptesting
// 8 s
//
// goversion: go1.20.1
// osversion: macOS 13.2.1
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/ptesting
// BenchmarkRand/math/rand.Uint32-10         	83335266	        13.93 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand/crypto/rand.Read_32-bit-10  	 3675235	       322.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand/runtime.fastrand_32-bit-10  	581979490	         2.085 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand/rand.Uint64-10              	88990515	        13.84 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand/FastRand.Uint64-10          	283184782	         4.234 ns/op	       0 B/op	       0 allocs/op
func BenchmarkRand(b *testing.B) {

	println(Versions())

	// math/rand.Uint32: thread-safe
	b.Run("math/rand.Uint32", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			rand.Uint32()
		}
	})

	// crypto/rand.Read: thread-safe
	b.Run("crypto/rand.Read 32-bit", func(b *testing.B) {
		var length = unsafe.Sizeof(uint32(1))
		var byts = make([]byte, length)
		b.ResetTimer()
		for iteration := 0; iteration < b.N; iteration++ {
			cryptorand.Read(byts)
		}
	})

	// runtime.fastrand: not thread-safe
	b.Run("runtime.fastrand 32-bit", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			fastrand()
		}
	})

	// math/rand.Uint64: thread-safe
	b.Run("rand.Uint64", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			rand.Uint64()
		}
	})

	// FastRand: not thread-safe
	b.Run("FastRand.Uint64", func(b *testing.B) {
		f := NewFastRand()
		for iteration := 0; iteration < b.N; iteration++ {
			f.Uint64()
		}
	})
}
