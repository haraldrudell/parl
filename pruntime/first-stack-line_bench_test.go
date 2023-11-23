/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import "testing"

// FirstStackLine is 2,020 ns/op on 2021 M1 Max
//   - 1.7 parallel mutex Lock/Unlock 1,178 wall-ns/op
//
// 231123 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkFirstStackLine$ github.com/haraldrudell/parl/pruntime
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl/pruntime
// BenchmarkFirstStackLine-10    	  590110	      2020 ns/op	     136 B/op	       2 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl/pruntime	1.394s
func BenchmarkFirstStackLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FirstStackLine()
	}
}
