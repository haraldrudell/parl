/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

// InvokeIf is a deferrable function invoking its function argument when:
//   - the pointer tp is non-nil and the function fn is non-nil
//   - what tp points to is not a T zero-value
//
// Usage:
//
//	someFlag := false
//	defer InvokeIf(&someFlag, someFunction)
//	…
//	someFlag = someValue
func InvokeIf[T comparable](tp *T, fn func()) {
	var zeroValueT T
	if tp == nil || fn == nil || *tp == zeroValueT {
		return
	}
	fn()
}
