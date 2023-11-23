/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"testing"
)

func TestStackTrace(t *testing.T) {
	// t.Fail()

	// var s = "(" + string(StackTrace()) + ")"
	// var s1 = strings.ReplaceAll(s, "\x20", "␠")
	// s1 = strings.ReplaceAll(s1, "\t", "␉")

	// // /opt/sw/parl/pruntime/stack-trace_test.go:17: stack trace:
	// // (goroutine␠20␠[running]:
	// // 	github.com/haraldrudell/parl/pruntime.StackTrace()
	// // 	␉/opt/sw/parl/pruntime/stack-trace.go:26␠+0x50
	// // 	github.com/haraldrudell/parl/pruntime.TestStackTrace(0x14000082ea0)
	// // 	␉/opt/sw/parl/pruntime/stack-trace_test.go:16␠+0x28
	// // 	testing.tRunner(0x14000082ea0,␠0x1022ec4c8)
	// // 	␉/opt/homebrew/Cellar/go/1.21.4/libexec/src/testing/testing.go:1595␠+0xe8
	// // 	created␠by␠testing.(*T).Run␠in␠goroutine␠1
	// // 	␉/opt/homebrew/Cellar/go/1.21.4/libexec/src/testing/testing.go:1648␠+0x33c
	// // 	)
	// t.Logf("stack trace:\n%s", s1)
}
