/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// onceDoErr provides a wrapper function for sync.once.Do
// that executes a function returning an error that may panic
//   - onceDoErr is a tuple value-type
//   - — if used as local variable, function argument or return value,
//     no allocation takes place
//   - — taking its address causes allocation
type onceDoErr struct {
	doErrFuncArgument func() (err error)
	// DoErr return value pointers
	didOnce *bool
	isPanic *bool
	errp    *error
	// pointer to observable parl.Once
	//	- has isDone and result fields
	*Once
}

// invokeDoErrFuncArgument is behind o.once
func (d *onceDoErr) invokeDoErrFuncArgument() {
	// indicate that the once did execute: last action
	defer d.isDone.Store(true)
	var result onceDoErrResult
	defer d.saveResult(&result)
	defer RecoverErr(func() DA { return A() }, &result.err, &result.isPanic)

	result.err = d.doErrFuncArgument()
}

// saveResult stores the outcome of a DoErr invocation
func (d *onceDoErr) saveResult(result *onceDoErrResult) {

	// store result in parl.Once
	d.result.Store(result)

	// update DoErr return values
	*d.didOnce = true
	*d.isPanic = result.isPanic
	*d.errp = result.err
}
