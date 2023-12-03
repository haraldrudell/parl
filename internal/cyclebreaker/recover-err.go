/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

// RecoverErr recovers panic using deferred annotation
//   - signature is error pointer and a possible isPanic pointer
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
