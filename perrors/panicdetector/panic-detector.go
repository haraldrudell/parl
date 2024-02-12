/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package panicdetector

const (
	// the fully-qualified function name searched for in a [runtime.Stack] value
	runtimeGopanic = "panic"
)

// panicDetector holds the function names searched for in
// [runtime.Stack]
type panicDetector struct {
	runtimeDeferInvokerLocation  string
	runtimePanicFunctionLocation string
}

// panicDetectorOne is a static value facilitating panic detection for the Go runtime
//   - created during package initialization therefore thread-safe
//   - must support current and previous Go versions,
//   - as of 240211: go1.21.0–go1.22.0
//   - used by [panicdetector.Indices]
//   - in a stack trace obtained from [runtime.Stack]:
//   - — function: panic
//   - — file: /opt/homebrew/Cellar/go/1.21.7/libexec/src/runtime/panic.go
//   - — additional Go runtime stack frames are not returned
//     - — the oldest stack frame is the first user function, ie.
//     main function or function provided in go statement
//   - in a frame obtained from [runtime.Callers]:
//   - — [runtime.Frame.Function]: runtime.gopanic
//   - — [runtime.Frame.File]: /opt/homebrew/Cellar/go/1.21.7/libexec/src/runtime/panic.go
//   - — several Go runtime stack frames are returned
//     - — the oldest stack frame is the Go exit function
var panicDetectorOne = &panicDetector{
	runtimeDeferInvokerLocation:  runtimeGopanic,
	runtimePanicFunctionLocation: runtimeGopanic,
}

// PanicDetectorValues is used by [whynotpanic.WhyNotPanic]
// to explain panic detection
func PanicDetectorValues() (deferS, panicS string) {
	deferS = panicDetectorOne.runtimeDeferInvokerLocation
	panicS = panicDetectorOne.runtimePanicFunctionLocation
	return
}
