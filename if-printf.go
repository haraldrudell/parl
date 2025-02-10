/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// PrintfFunc is the signature for a printf-style function.
// This signature is implemented by:
//   - parl.Sprintf
//   - parl.Out parl.Outw parl.Log parl.Logw parl.Console parl.Consolew
//     parl.Info parl.Debug parl.D parl.NoPrint
//   - plog.(*LogInstance) similar methods
//   - perrors.Errorf perrors.ErrorfPF
//   - fmt.Printf fmt.Sprintf fmt.Errorf
//   - pterm.(*StatusTerminal).Log pterm.(*StatusTerminal).LogTimeStamp
//
// and compatible functions
type PrintfFunc func(format string, a ...any)

// Logger is a generic PrintfFunc value
type Logger interface{ Log(format string, a ...any) }
