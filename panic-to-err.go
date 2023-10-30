/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

// PanicToErr recovers active panic, aggregating errors in errp
//   - errp cannot be nil
//   - if isPanic is non-nil and active panic, it is set to true
//
// sample error message, including message in the panic value and the code line
// causing the panic:
//
//	recover from panic: message: “runtime error: invalid memory address or nil pointer dereference” at parl.panicFunction()-panic-to-err_test.go:96
//
// Usage:
//
//	func someFunc() (isPanic bool, err error) {
//	  defer parl.PanicToErr(&err, &isPanic)
//
//	func someGoroutine(g parl.Go) {
//	  var err error
//	  defer g.Register().Done(&err)
//	  defer parl.PanicToErr(&err)
func PanicToErr(errp *error, isPanic ...*bool) {
	if errp == nil {
		panic(perrors.NewPF("errp cannot be nil"))
	}

	// if no panic, noop
	//	- recover invocation must be directly in the PanicToErr function
	var panicValue = recover()
	if panicValue == nil {
		return // no panic active return
	}

	// set isPanic if non-nil
	if len(isPanic) > 0 {
		if isPanicp := isPanic[0]; isPanicp != nil {
			*isPanicp = true
		}
	}

	// append panic to *errp
	//	- for panic detection to work there needs to be a stack frame after runtime
	//	- because PanicToErr is invoked directly by the runtime, eg. pruntime.gopanic,
	//		the PanicToErr stack frame is included, therefore 0 argument
	//		to processRecover
	*errp = perrors.AppendError(*errp,
		processRecover("recover from panic: message:", panicValue, 0),
	)
}
