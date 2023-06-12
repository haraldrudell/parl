/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	debugThunkFrames = 2 // DebugThunk + IsThisDebugN
)

// DebugThunk is similar to parl.Debug but arguments are only resolved when
// debug is true, ie. the arguments will actually be printed
//   - the argument can be a function literal invoking parl.Sprintf
//
// Usage:
//
//	parl.DebugThunk(func() string { return parl.Sprintf("a: %d", 3)}) })
func DebugThunk(argThunk func() string) {
	if !IsThisDebugN(debugThunkFrames) {
		return
	}
	stderrLogger.Log(argThunk())
}
