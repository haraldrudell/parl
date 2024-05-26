/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Callbacker accepts asynchronously provided values
type Callbacker[T any] interface {
	// Callback is an asynchronous return of a value
	//	- Callback typically needs to be thread-safe
	Callback(value T)
}

// CallbackerErr accepts asynchronously provided values
// and can fail
type CallbackerErr[T any] interface {
	// Callback is an asynchronous return of a value that can fail
	//	- Callback typically needs to be thread-safe
	Callback(value T) (err error)
}

var NoCallbackerNoArg CallbackerNoArg

// CallbackerNoArg receives asynchronous event notifications
type CallbackerNoArg interface {
	// Callback is an asynchronous event notifier
	//	- Callback typically needs to be thread-safe
	Callback()
}

var NoCallbackerNoArgErr CallbackerNoArgErr

// CallbackerNoArgErr receives asynchronous event notifications
// and can fail
type CallbackerNoArgErr interface {
	// Callback is an asynchronous event notifier that can fail
	//	- Callback typically needs to be thread-safe
	Callback() (err error)
}
