/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// TFunc is a function that returns value, err and may panic
type TFunc[T any] func() (value T, err error)

// TResult is a value-container for value, isPanic and error
type TResult[T any] struct {
	Value   T
	IsPanic bool
	Err     error
}

// NewTResult3 creates a TResult from pointers at the time values are available
//   - value is considered valid if errp is nil or *errp is nil
//   - any arguments may be nil
func NewTResult3[T any](value *T, isPanic *bool, errp *error) (tResult *TResult[T]) {
	var result TResult[T]
	tResult = &result
	if errp != nil {
		if err := *errp; err != nil {
			result.Err = err
			if isPanic != nil {
				result.IsPanic = *isPanic
			}
		}
	}
	if result.Err == nil && value != nil {
		result.Value = *value
	}
	return
}

// NewTResult creates a result container
//   - if tFunc is present, it is invoked prior to returning storing its result
//   - recovers tFunc panic
func NewTResult[T any](tFunc ...TFunc[T]) (tResult *TResult[T]) {

	// create result object
	var t TResult[T]
	tResult = &t

	// check if tFunc is available
	var f TFunc[T]
	if len(tFunc) > 0 {
		f = tFunc[0]
	}
	if f == nil {
		return // tFunc not present return
	}

	// execute tFunc
	defer RecoverErr(func() DA { return A() }, &t.Err, &t.IsPanic)

	t.Value, t.Err = f()

	return // tFunc completed return
}
