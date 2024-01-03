/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import "github.com/haraldrudell/parl/perrors"

const (
	panicToErrAnnotation = "recover from panic: message:"
)

// PanicToErr recovers active panic, aggregating errors in errp
//   - PanicToErr does not provide enclosing function. For that,
//     use RecoverErr: “recover from panic in pack.Func…”
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
		panic(NilError("errp"))
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
	//	- for panic detector to work there needs to be at least one stack frame after runtime’s
	//		panic handler
	//	- because PanicToErr is invoked directly by the runtime, possibly runtime.gopanic,
	//		the PanicToErr stack frame must be included.
	//		Therefore, 0 argument to processRecover
	*errp = perrors.AppendError(*errp, processRecoverValue(panicToErrAnnotation, panicValue, 0))
}
