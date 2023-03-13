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

// go test -run=^# -bench=BenchmarkRand32 ./ptesting
// 5 s
//
// goversion: go1.20.1
// osversion: macOS 13.2.1
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/ptesting
// BenchmarkRand32/math/rand.Uint32-10         	84712857	        14.19 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand32/crypto/rand.Read_32-bit-10  	 3601936	       329.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand32/ptesting.Uint32-10          	566568112	         2.162 ns/op	       0 B/op	       0 allocs/op
func BenchmarkRand32(b *testing.B) {

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

	// runtime.fastrand 32-bit: thread-safe
	b.Run("ptesting.Uint32", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			Uint32()
		}
	})
}

// go test -run=^# -bench=BenchmarkRand64 ./ptesting
// 5 s
//
// goversion: go1.20.1
// osversion: macOS 13.2.1
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/ptesting
// BenchmarkRand64/math/rand.Uint64-10         	84433459	        14.01 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand64/crypto/rand.Read_64-bit-10  	 3783060	       323.8 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRand64/ptesting.Uint64-10          	277811217	         4.289 ns/op	       0 B/op	       0 allocs/op
func BenchmarkRand64(b *testing.B) {

	println(Versions())

	// math/rand.Uint64: thread-safe
	b.Run("math/rand.Uint64", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			rand.Uint64()
		}
	})

	// crypto/rand.Read: thread-safe
	b.Run("crypto/rand.Read 64-bit", func(b *testing.B) {
		var length = unsafe.Sizeof(uint64(1))
		var byts = make([]byte, length)
		b.ResetTimer()
		for iteration := 0; iteration < b.N; iteration++ {
			cryptorand.Read(byts)
		}
	})

	// runtime.fastrand 64-bit: thread-safe
	b.Run("ptesting.Uint64", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			Uint64()
		}
	})
}
