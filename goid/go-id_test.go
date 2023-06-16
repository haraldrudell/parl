/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import (
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
)

func TestGoID(t *testing.T) {
	runtGoroutinePrefix := "goroutine "

	// get expected ThreadID
	var expectedID parl.ThreadID
	s := strings.TrimPrefix(string(debug.Stack()), runtGoroutinePrefix)
	if index := strings.Index(s, "\x20"); index != -1 {
		expectedID = parl.ThreadID(s[:index])
	} else {
		t.Error("debug.Stack failed")
	}

	actual := GoID()
	if actual != expectedID {
		t.Errorf("GoRoutineID bad: %q expected %q", actual, expectedID)
	}
}

/*
2022-04-08 host: c66 go version: go1.18
the selected number of iterations: 480,165
execution time: 2.364 μs
the code allocated 1,664 bytes i 2 allocations

# A function call is 0.3215 ns

goos: darwin
goarch: arm64
pkg: github.com/haraldrudell/parl/goid
BenchmarkGoID-10    	  480165	      2364 ns/op	    1664 B/op	       2 allocs/op
*/
func BenchmarkData(b *testing.B) {
	// a global variable is undefined every time
	// os.MkdirTemp returns a different directory every time
	if hostname, err := os.Hostname(); err != nil {
		b.Errorf("os.Hostname FAIL: %v", err)
	} else {
		if index := strings.Index(hostname, "."); index != -1 {
			hostname = hostname[:index]
		}
		today := time.Now().Format("2006-01-02")
		b.Logf("%s host: %s go version: %s", today, hostname, runtime.Version())
	}
	for i := 0; i < b.N; i++ {
		GoID()
	}
}

func BenchmarkGoID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GoID()
	}
}

func BenchmarkCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		someFunc()
	}
}

func someFunc() {}
