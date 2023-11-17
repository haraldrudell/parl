/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// when the iterator function receives this value, it means cancel
const FunctionIteratorCancel int = -1

// IteratorFunction is the signature used by NewFunctionIterator
//   - if isCancel true, it means this is the last invocation of IteratorFunction and
//     IteratorFunction should release any resources.
//     Any returned value is not used
//   - IteratorFunction signals end of values by returning parl.ErrEndCallbacks.
//     Any returned value is not used
//   - if IteratorFunction returns error, it will not be invoked again.
//     Any returned value is not used
//   - IteratorFunction must be thread-safe
//   - IteratorFunction is invoked by at most one thread at a time
type IteratorFunction[T any] func(isCancel bool) (value T, err error)

// ConverterFunction is the signature used by NewConverterIterator
//   - ConverterFunction receives a key and returns the corresponding value.
//   - if isCancel true, it means this is the last invocation of ConverterFunction and
//     ConverterFunction should release any resources.
//     Any returned value is not used
//   - ConverterFunction signals end of values by returning parl.ErrEndCallbacks.
//     Any returned value is not used
//   - if ConverterFunction returns error, it will not be invoked again.
//     Any returned value is not used
//   - ConverterFunction must be thread-safe
//   - ConverterFunction is invoked by at most one thread at a time
type ConverterFunction[K any, V any] func(key K, isCancel bool) (value V, err error)

// InvokeFunc is the signature used by NewBaseIterator
//   - isCancel is a request to cancel iteration. No more invocations
//     will occur. Iterator should release resources
//   - value is valid if err is nil and isCancel was false and didCancel is false.
//     otherwise, value is ignored
//   - if err is non-nil an error occurred.
//     parl.ErrEndCallbacks allows InvokeFunc to signal end of iteration values.
//     parl.ErrEndCallbacks error is ignored.
//   - if isCancel was false and didCancel is true, IncokeFunc took the
//     initiative to cancel the iteration. No more invocations will occur
//   - isPanic indicates that err is the result of a panic
type InvokeFunc[T any] func(isCancel bool) (
	value T,
	didCancel bool,
	isPanic bool,
	err error,
)
