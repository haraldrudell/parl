/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Done declares an object with a Done method
//   - The Done interface is similar to a done callback function value
//   - — passing a method as function value causes allocation
//   - — such function values create difficult-to-follow stack traces
//   - The Done interface implemented by eg. [sync.WaitGroup.Done]
//   - callbacks are used when:
//   - — invoked repeatedly
//   - — invoked by a subthread or asynchronous method
//   - a thread-safe callback may save a listening thread and a lock
//   - alternatives to callback are:
//   - — blocking return or
//   - — lock-featuring synchronization mechanics like chan costing a thread
type Done interface {
	// Done signals that some task has completed.
	Done()
}
