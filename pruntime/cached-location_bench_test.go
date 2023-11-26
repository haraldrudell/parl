/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"testing"
)

// subsequent is 0.6833 ns -99.93%
//   - initial is 1,004 ns
//   - — 0.85 parallel mutex Lock/Unlock 1,178 wall-ns/op
//
// 231126 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCachedLocationSuite$ github.com/haraldrudell/parl/pruntime
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/pruntime
// BenchmarkCachedLocationSuite/Init-10         	 1165704	      1004 ns/op	     456 B/op	       8 allocs/op
// BenchmarkCachedLocationSuite/Cached-10       	1000000000	         0.6833 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl/pruntime	2.969s
func BenchmarkCachedLocationSuite(b *testing.B) {
	var benchs = []struct {
		name  string
		bench func(b *testing.B)
	}{
		{"Init", BenchmarkCachedLocationInit},
		{"Cached", BenchmarkCachedLocation},
	}
	for _, bench := range benchs {
		b.Run(bench.name, bench.bench)
	}

}

func BenchmarkCachedLocationInit(b *testing.B) {
	var c CachedLocation
	for i := 0; i < b.N; i++ {
		c = CachedLocation{}
		c.PackFunc()
	}
}

func BenchmarkCachedLocation(b *testing.B) {
	var c CachedLocation
	c.PackFunc()
	for i := 0; i < b.N; i++ {
		c.PackFunc()
	}
}
