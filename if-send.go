/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Send declares an object with a Send method
//   - the Send interface is similar to a callback function value
//   - — passing a method as function value causes allocation
//   - — such function values create difficult-to-follow stack traces
//   - the Send interface is implemented by eg. [github.com/haraldrudell/parl.NBChan]
//   - Send is intended to provide a trouble-free value sink
//     transferring data to other threads
type Send[T any] interface {
	// Send is typically a thread-safe non-blocking panic-free error-free
	// send hand-off of a single value to another thread implemented as a method
	//	- Send replaces Go channel send operation
	Send(value T)
}
