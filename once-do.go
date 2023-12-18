/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// onceDo provides a wrapper method
// that invokes doFuncArgument and then sets isDone to true
//   - onceDo is a tuple value-type
//   - — if used as local variable, function argument or return value,
//     no allocation tajkes place
//   - — taking its address causes allocation
type onceDo struct {
	doFuncArgument func()
	// pointer to observable parl.Once
	//	- has isDone field
	*Once
}

// invokeF is behind o.once
//   - after doFuncArgument invocation, it sets isDone to true
//   - isDone provides observability
func (d *onceDo) invokeF() {
	defer d.isDone.Store(true)

	d.doFuncArgument()
}
