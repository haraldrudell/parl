/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"testing"
)

// PackFunc is 467.1 ns on 2021 M1 Max
//   - 783 64-bit parallel atomic reads 0.5962 wall-ns/op
//   - 0.69 Kimap reads 680.5 wall-ns/op
//   - 0.40 parallel mutex Lock/Unlock 1,178 wall-ns/op
//
// 231123 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkPackFunc$ github.com/haraldrudell/parl/pruntime
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/pruntime
// BenchmarkPackFunc-10    	 2627085	       467.1 ns/op	     280 B/op	       3 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl/pruntime	1.983s
func BenchmarkPackFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCodeLocation(0).PackFunc()
	}
}
