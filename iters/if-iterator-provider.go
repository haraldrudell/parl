/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

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

type SimpleIteratorFunc[T any] func() (value T, hasValue bool)

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

type SimpleConverter[K any, V any] func(key K) (value V)

// IteratorAction is a delegated request from [iters.BaseIterator]
//   - isCancel true requests cancel of iteration.
//     No further invocations will occur.
//     The iterator should release resources.
//     The iterator may return an error
//   - otherwise, the iterator can:
//   - — return the next value
//   - — return an error. No further invocations will occur
//   - — return err == parl.ErrEndCallbacks requesting an end to iterations.
//     No further invocations will occur
//     ErrEndCallbacks error is not returned to the consumer
//   - the returned value is used if:
//   - — returned err is nil and
//   - — provided isCancel was false and
//   - — returned didCancel is false
//   - isPanic indicates that err is the result of a panic.
//     isPanic is only used if err is non-nil
type IteratorAction[T any] func(isCancel bool) (
	value T,
	isPanic bool,
	err error,
)
