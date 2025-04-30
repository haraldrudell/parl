/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// DoDone recovers panic with defer location, validates any errors and executes completion action
//   - single-defer invocation
//   - deferredLocation: function literal providing defer location
//   - doner: deferred completion action
//   - errp: pointer to function-scoped error value
//   - isPanic: optional panic flag indicating process-terminating error
//   - parl recover options:
//   - — [DoDone]: function-scoped error with defer location and panic-flag integrated with completion action
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
//	func goFunction(g parl.Go) {
//	  var err error
//	  defer parl.DoDone(func() parl.DA { return parl.A() }, g.Register(), &err)
//
//	func taskDoer(task parly.Task) (isPanic bool) {
//	  var err error
//	  defer parl.DoDone(func() parl.DA { return parl.A() }, task, &err, &isPanic)
func DoDone(deferredLocation func() DA, doner Doner, errp *error, isPanic ...*bool) {

	// isPanicp is pointer to isPanic flag or nil
	var isPanicp *bool
	if len(isPanic) > 0 {
		isPanicp = isPanic[0]
	}

	// collect panic and ensure error stack trace
	doRecovery(noAnnotation, deferredLocation, errp, recoverOnErrrorNone, isPanicp, recover())

	// execute Done: exiting thread, completing thread-independent task or
	// other completion action
	doner.Done(errp)
}
