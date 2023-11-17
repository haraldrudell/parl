/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package prand

import (
	cryptorand "crypto/rand"
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/haraldrudell/parl/ptesting"
)

// How to generate pseudo-random numbers efficiently?
//	- use fastrand 2.092 ns
//	- create thread-local math/rand.Rand and then use Uint32: 2.240 ns
//	- pick a slow option

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

	println(ptesting.Versions())

	// math/rand.Uint32: thread-safe
	b.Run("math/rand.Uint32", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			rand.Uint32()
		}
	})

	// math/rand.Uint32: thread-safe
	b.Run("math/rand.Rand.Uint32", func(b *testing.B) {
		// random generator with thread-local seed
		var r = rand.New(rand.NewSource(0))
		for iteration := 0; iteration < b.N; iteration++ {
			r.Uint32()
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

	println(ptesting.Versions())

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

// thread-local math/rand.Rand.Uint32 pseudo-random number on 2021 M1 Max is 2.240 ns
//   - thread must have Rand available
//
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkNewRand$ github.com/haraldrudell/parl/prand
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/prand
// BenchmarkNewRand-10    	53342386	        22.40 ns/op	         2.240 wall-ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl/prand	2.234s
func BenchmarkNewRand(b *testing.B) {
	const uint32Count = 10
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Uint32() // 1
		r.Uint32() // 2
		r.Uint32() // 3
		r.Uint32() // 4
		r.Uint32() // 5
		r.Uint32() // 6
		r.Uint32() // 7
		r.Uint32() // 8
		r.Uint32() // 9
		r.Uint32() // 10
	}
	b.StopTimer()
	// elapsed is duration of 10 fastrand invocations. Unit ns, int64
	var elapsed = b.Elapsed()
	// nUnit is wall-time nanoseconds per fastrand invocation
	//	- float64, unit ns
	var nUnit float64 = //
	float64(elapsed) /
		float64(b.N) /
		uint32Count
	b.ReportMetric(nUnit, "wall-ns/op")
}

// time.Now().UnixNano() as 64-bit pseudo-random number on 2021 M1 Max is 38.00 ns
//
// 231114 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkTimeNow$ github.com/haraldrudell/parl/prand
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/prand
// BenchmarkTimeNow-10    	31434660	        38.00 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl/prand	2.325s
func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now().UnixNano()
	}
}

// fastrand 32-bit pseudo-random number on 2021 M1 Max is 2.092 ns
//
// 231114 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkFastRand$ github.com/haraldrudell/parl/prand
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/prand
// BenchmarkFastRand-10    	57573570	        20.92 ns/op	         2.092 wall-ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl/prand	2.288s
func BenchmarkFastRand(b *testing.B) {
	const fastRandCount = 10
	for i := 0; i < b.N; i++ {
		// fastrand invocation is 2.039 ns which is too short to be tested by itself
		//	- at least 4 ns of measured task is required for metric quality
		//	- 10 fastrand is 21.10 ns
		fastrand() // 1
		fastrand() // 2
		fastrand() // 3
		fastrand() // 4
		fastrand() // 5
		fastrand() // 6
		fastrand() // 7
		fastrand() // 8
		fastrand() // 9
		fastrand() // 10
	}
	b.StopTimer()
	// elapsed is duration of 10 fastrand invocations. Unit ns, int64
	var elapsed = b.Elapsed()
	// nUnit is wall-time nanoseconds per fastrand invocation
	//	- float64, unit ns
	var nUnit float64 = //
	float64(elapsed) /
		float64(b.N) /
		fastRandCount
	b.ReportMetric(nUnit, "wall-ns/op")
}
