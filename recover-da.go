/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pruntime"

const (
	// counts the frames in [parl.A]
	parlAFrames = 1
)

// DA is the value returned by a deferred code location function
type DA *pruntime.CodeLocation

// A is a thunk returning a deferred code location
func A() DA { return pruntime.NewCodeLocation(parlAFrames) }

// RecoverDA recovers panic using deferred annotation
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.RecoverDA(func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func RecoverDA(deferredLocation func() DA, errp *error, onError OnError) {
	doRecovery("", deferredLocation, errp, onError, recover2OnErrrorOnce, noIsPanic, recover())
}

// RecoverErr recovers panic using deferred annotation
//
// Usage:
//
//	func someFunc() (isPanic bool, err error) {
//	  defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err, &isPanic)
func RecoverErr(deferredLocation func() DA, errp *error, isPanic ...*bool) {
	var isPanicp *bool
	if len(isPanic) > 0 {
		isPanicp = isPanic[0]
	}
	doRecovery("", deferredLocation, errp, NoOnError, recover2OnErrrorOnce, isPanicp, recover())
}

// RecoverDA2 recovers panic using deferred annotation
//
// Usage:
//
//	func someFunc() (err error) {
//	  defer parl.RecoverDA2(func() parl.DA { return parl.A() }, &err, parl.NoOnError)
func RecoverDA2(deferredLocation func() DA, errp *error, onError OnError) {
	doRecovery("", deferredLocation, errp, onError, recover2OnErrrorMultiple, noIsPanic, recover())
}
