/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// RecoverErr recovers panic using deferred annotation
//   - signature is error pointer and a possible isPanic pointer
//   - parl recover options:
//   - — [RecoverErr]: aggregates to error pointer with enclosing function location, optional panic flag
//   - — [Recover]: aggregates to error pointer with enclosing function location, optional single-invocation [parl.ErrorSink]
//   - — [Recover2]: aggregates to error pointer with enclosing function location, optional multiple-invocation [parl.ErrorSink]
//   - — [RecoverAnnotation]: aggregates to error pointer with fixed-string annotation, optional single-invocation [parl.ErrorSink]
//   - — [PanicToErr]: aggregates to error pointer with generic annotation, optional panic flag
//   - — preferrably: RecoverErr, Recover or Recover2 should be used to provide the package and function name
//     of the enclosing function for the defer-statement that invoked recover
//   - — PanicToErr and RecoverAnnotation cannot provide where in the stack trace recover was invoked
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
	doRecovery(noAnnotation, deferredLocation, errp, recoverOnErrrorNone, isPanicp, recover())
}
