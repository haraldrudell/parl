/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type TFunc[T any] func() (value T, err error)

type TResult[T any] struct {
	Value   T
	IsPanic bool
	Err     error
}

// RecoverInvocationPanicT invokes resolver, recover panics and populates v
func NewTResult[T any](tFunc TFunc[T]) (tResult *TResult[T]) {
	var t = TResult[T]{IsPanic: true}
	tResult = &t
	defer PanicToErr(&t.Err)

	t.Value, t.Err = tFunc()
	t.IsPanic = false
	return
}
