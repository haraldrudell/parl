/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	debugThunkFrames = 2 // DebugThunk + IsThisDebugN
)

// OnDebug is similar to parl.Debug but arguments are only resolved when
// debug is true, ie. when arguments should actually be printed
//   - the argument can be a function literal invoking parl.Sprintf
//
// Usage:
//
//	var x int
//	parl.OnDebug(func() string { return parl.Sprintf("before: %d", x)})
func OnDebug(invokedIfDebug func() string) {
	if !IsThisDebugN(debugThunkFrames) {
		return
	}
	stderrLogger.Log(invokedIfDebug())
}
