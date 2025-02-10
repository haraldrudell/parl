/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Send declares an object with a Send method
//   - Send is intended to provide a trouble-free value sink
//     transferring data to other threads
//   - — typically thread-safe non-blocking panic-free error-free
//   - the Send interface replaces channel send operation
//   - — the object can freely implement Send
//   - — may avoid panics, blocks and difficulties with channel send
//   - the Send interface replaces a callback function value
//   - — passing a method as function value in Go causes 18 ns allocation
//   - — such function values may create difficult-to-follow stack traces
//   - the Send interface is implemented by eg. [github.com/haraldrudell/parl.NBChan]
type Send[T any] interface {
	// Send is typically a thread-safe non-blocking panic-free error-free
	// send hand-off of a single value to another thread implemented as a method
	//	- Send replaces Go channel send operation
	Send(value T)
}
