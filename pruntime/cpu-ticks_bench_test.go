/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"math/rand"
	"testing"
	"time"
)

// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkCpuTicks$ github.com/haraldrudell/parl/pruntime
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/pruntime
// BenchmarkCpuTicks-10    	74907470	        16.04 ns/op	       0 B/op	       0 allocs/op
func BenchmarkCpuTicks(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CpuTicks()
	}
}

func BenchmarkSingleThreadRandomInt(b *testing.B) {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Intn(3)
	}
}
