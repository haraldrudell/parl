/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// ValueSink allows a thread to submit data to other threads
//   - implemented by [github.com/haraldrudell/parl.AwaitableSlice]
type ValueSink[T any] interface {
	// Send submits single value
	Send(value T)
	// SendSlice submits any number of values
	SendSlice(values []T)
	// EmptyCh provides close-like behavior
	//	- this thread invoking EmptyCh() signals end of values
	//	- other threads invoking EmptyCh(parl.CloseAwaiter) awaits:
	//	- — the EmptyCh() invocation and
	//	- — the AwaitableSlice becoming empty
	EmptyCh(doNotInitialize ...bool) (ch AwaitableCh)
}

// DataSink allows a thread to submit data to other threads
//   - implemented by [github.com/haraldrudell/parl.AwaitableSlice]
type ValueSource[T any] interface {
	// Get obtains a single value
	Get() (value T, hasValue bool)
	// GetSlice obtains values by the slice
	//	- nil slice means source is currently empty
	GetSlice() (values []T)
	// GetAll receives a combined slice of all values
	//	- may cause allocation
	GetAll() (values []T)
	// Init allows for AwaitableSlice to be used in a for clause
	//   - returns zero-value for a short variable declaration in
	//     a for init statement
	//
	// Usage:
	//
	//	var a AwaitableSlice[…] = …
	//	for value := a.Init(); a.Condition(&value); {
	//	  // process received value
	//	}
	//	// the AwaitableSlice closed
	Init() (value T)
	// Condition allows for AwaitableSlice to be used in a for clause
	//   - updates a value variable and returns whether values are present
	//
	// Usage:
	//
	//	var a AwaitableSlice[…] = …
	//	for value := a.Init(); a.Condition(&value); {
	//	  // process received value
	//	}
	//	// the AwaitableSlice closed
	Condition(valuep *T) (hasValue bool)
	// DataWaitCh returns an awaitable closing channel that
	//	- is closed if data is available or
	//	- closes once data becomes available
	DataWaitCh() (ch AwaitableCh)
	// EmptyCh provides close-like behavior
	//	- this thread invoking EmptyCh(parl.CloseAwaiter) returns a
	//		channel that closes on stream completion:
	//	- — EmptyCh() invocation with no value by other thread and
	//	- — the AwaitableSlice becoming empty
	EmptyCh(doNotInitialize ...bool) (ch AwaitableCh)
}
