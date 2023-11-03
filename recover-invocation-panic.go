/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// RecoverInvocationPanic is intended to wrap callback invocations in the callee in order to
// recover from panics in the callback function.
// when an error occurs, perrors.AppendError appends the callback error to *errp.
// if fn is nil, a recovered panic results.
// if errp is nil, a panic is thrown, can be check with:
//
//	if errors.Is(err, parl.ErrNilValue) …
func RecoverInvocationPanic(fn func(), errp *error) {
	if errp == nil {
		panic(NilError("errp"))
	}
	defer PanicToErr(errp)

	fn()
}

func RecoverInvocationPanicErr(fn func() (err error)) (isPanic bool, err error) {
	defer PanicToErr(&err, &isPanic)

	err = fn()

	return
}
