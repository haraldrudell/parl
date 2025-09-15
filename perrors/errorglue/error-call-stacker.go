/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "github.com/haraldrudell/parl/pruntime"

// ErrorCallStacker enrichens an error with a stack trace of code locations
type ErrorCallStacker interface {
	StackTrace() (stack pruntime.Stack)
}
