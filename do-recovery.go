/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// DoRecovery facilitates custom recover functions by providing access to doRecovery
//   - deferredLocation: provides initial defer location
//   - recoverValue: provides panic recovery
//   - errp: provided error value
//   - purpose is to defer only one function
//
// Usage:
//
//	func (s *someStruct) goFunction(g parl.Go) {
//	  var err error
//	  defer s.customRecovery(func() parl.DA { return parl.A() }, g.Register(), &err)
//
//	func (s *someStruct) customRecovery(deferredLocation func() DA, g parl.ErrorDoner, errp *error) {
//	  parl.DoRecovery(deferredLocation, recover(), errp)
//	  g.Done(errp)
func DoRecovery(deferredLocation func() DA, recoverValue any, errp *error) {
	// collect panic and ensure error stack trace
	doRecovery(noAnnotation, deferredLocation, errp, recoverOnErrrorNone, noPanicp, recoverValue)
}

var (
	// noPanicp indicates no panic flag present
	noPanicp *bool
)
