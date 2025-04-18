/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"
	"iter"
)

//	- sink for values and slices
//	- closable sink for values and slices
//	- source for values and slices
//	- source for values and slices with all
//	- closable source for values and slices
//	- closable source for values and slices with all

// Sink is a stream-object receiving values
//   - no concept of close or drain
//   - similar to channel send operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock,
//     blocking or eventually completing
//   - no panic or error: error handling is separate
//   - flexible allocation strategy
//   - implemented by [AwaitableSlice]
//   - methods: Send SendSlice SendClone
type Sink[T any] interface {
	// Send sends a single value
	//	- Sink is responsible for any allocation
	Send(value T)
	// SendSlice sends a slice of values
	//	- Sink takes ownership of the slice
	//	- used for slice-allocation at the source with
	//		slice ownership forwarded through to the final endpoint
	SendSlice(values []T)
	// SendClone sends a clone of the slice
	//	- Sink is responsible for any allocation
	//	- used when the source owns a slice buffer
	//		whose content is being sent
	SendClone(values []T)
}

// ClosableSink is a stream-object receiving values
// featuring close-drain mechanic
//   - similar to channel send operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock,
//     blocking or eventually completing
//   - no panic or error: error handling is separate
//   - flexible allocation strategy
//   - implemented by [AwaitableSlice]
//   - methods: Send SendSlice SendClone CloseCh IsClosed
type ClosableSink[T any] interface {
	// Send SendSlice SendClone
	Sink[T]
	// CloseCh IsClosed
	Closable[T]
}

// Source1 is a source stream-object
// providing single value at a time
//   - Source1 is awaitable using closing-channel mechanic
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue
type Source1[T any] interface {
	// Get is non-blocking receive of single value
	//	- value: possibly received value
	//	- hasValue: true if a value was provided
	Get() (value T, hasValue bool)
	// DataWaitCh upon each invocation returns
	// closing-channel mechanic
	// that is closed or closes when data is available
	//
	// usage:
	//	for {
	//	  select {
	//	    case <-source1.DataWaitCh():
	//	      // competing with other threads for values
	//	      //	- may receive nothing
	//	      if value, hasValue = s.Get(); hasValue {
	//	        …
	DataWaitCh() (ch AwaitableCh)
	// AwaitValue awaits value or close, blocking until either event
	//	- value: possibly received value
	//	- hasValue: true if a value was provided
	//	- similar to [AwaitValue]
	AwaitValue() (value T, hasValue bool)
}

// Source is a source stream-object
// providing single value or owned slice at a time
//   - Source is awaitable using closing-channel mechanic
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice
type Source[T any] interface {
	// Get DataWaitCh AwaitValue
	Source1[T]
	// GetSlice returns a slice of values from the source
	// or nil
	//	- consumer takes ownership of the slice
	GetSlice() (values []T)
}

// AllSource is a stream-object source that can return all its values
//   - AllSource is awaitable using closing-channel mechanic
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice GetAll
type AllSource[T any] interface {
	// Get DataWaitCh AwaitValue GetSlice
	Source[T]
	// GetAll returns a slice of all values from the source
	// or nil
	//	- consumer takes ownership of the slice
	GetAll() (values []T)
}

// ClosableSource1 is a stream-object source that is
// awaitable, closable providing one value at a time
//   - ClosableSource1 is awaitable using closing-channel mechanic
//     and features close-drain
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue CloseCh IsClosed
type ClosableSource1[T any] interface {
	// Get DataWaitCh AwaitValue
	Source1[T]
	// CloseCh IsClosed
	Closable[T]
}

// IterableSource is a stream-source that can be used with
// for range statement
//   - IterableSource is awaitable using closing-channel mechanic
//     and features close-drain
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue CloseCh IsClosed Seq
type IterableSource[T any] interface {
	// Get DataWaitCh AwaitValue CloseCh IsClosed
	ClosableSource1[T]
	// Seq is an iterator over sequences of individual values.
	// When called as seq(yield), seq calls yield(v) for
	// each value v in the sequence, stopping early if yield returns false.
	Seq(yield func(value T) (keepGoing bool))
}

// IterableSource is a stream-source that can be used with
// for range statement
//   - IterableSource is awaitable using closing-channel mechanic
//     and features close-drain
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice GetAll CloseCh IsClosed Seq
type IterableAllSource[T any] interface {
	// Get DataWaitCh AwaitValue GetSlice GetAll CloseCh IsClosed
	ClosableAllSource[T]
	// Seq is an iterator over sequences of individual values.
	// When called as seq(yield), seq calls yield(v) for
	// each value v in the sequence, stopping early if yield returns false.
	Seq(yield func(value T) (keepGoing bool))
}

// IterableSource implements one-value iteration: [iter.Seq]
var _ = func(i IterableSource[int]) (s iter.Seq[int]) { return i.Seq }

// ClosableSource is a source stream-object
// providing single value or owned slice at a time
//   - ClosableSource is awaitable using closing-channel mechanic
//     and features close-drain
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice CloseCh IsClosed
type ClosableSource[T any] interface {
	// Get DataWaitCh AwaitValue GetSlice
	Source[T]
	// CloseCh IsClosed
	Closable[T]
}

// ClosableAllSource is a closable stream-object source that can
// return all its values
//   - ClosableAllSource is awaitable using closing-channel mechanic
//     and features close-drain
//   - similar to channel receive operation
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice GetAll CloseCh IsClosed
type ClosableAllSource[T any] interface {
	// Get DataWaitCh AwaitValue GetSlice GetAll
	AllSource[T]
	// CloseCh IsClosed
	Closable[T]
}

// SourceSink is a stream-object source and sink that can return
// all its values
//   - SourceSink is awaitable using closing-channel mechanic
//   - no concept of close or drain
//   - similar to channel send and receive operations
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - flexible allocation strategy
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice GetAll Send SendSlice SendClone
type SourceSink[T any] interface {
	// Get DataWaitCh AwaitValue GetSlice GetAll
	AllSource[T]
	// Send SendSlice SendClone
	Sink[T]
}

// ClosableSourceSink is a closable stream-object source and sink that can return
// all its values
//   - ClosableSourceSink is awaitable using closing-channel mechanic
//     and features close-drain
//   - similar to channel send and receive operations
//   - a stream is a thread-safe awaitable queue similar to
//     a channel with unbound buffer
//   - implementation may be intra-thread,
//     or thread-safe via atomic or lock
//   - no panic or error: error handling is separate
//   - flexible allocation strategy
//   - implemented by [AwaitableSlice]
//   - methods: Get DataWaitCh AwaitValue GetSlice GetAll Send SendSlice SendClone CloseCh IsClosed
type ClosableSourceSink[T any] interface {
	// Get DataWaitCh AwaitValue GetSlice GetAll Send SendSlice SendClone
	SourceSink[T]
	// CloseCh IsClosed
	Closable[T]
}

// Closable is a stream-object [Sink] or [Source] that is closable
//   - a component interface not intended for consumer use
//   - present in [ClosableSink]
//     [ClosableSource1] [ClosableSource] [ClosableAllSource] [IterableSource]
//     [ClosableSourceSink]
//   - may be used to determine if a stream is closable
//   - implemented by [AwaitableSlice]
//   - methods: CloseCh IsClosed
type Closable[T any] interface {
	// CloseCh returns a channel that is closed or closes
	// upon the stream becoming empty
	//	- ch: used to await close and drain
	//	- CloseCh always returns the same channel value
	//	- close state is separate from value flow:
	//		a closed sink will still receive values.
	//		A closed source that is emptied may become unempty again.
	//	- after a close invocation when the stream is empty,
	//		the stream is marked as closed.
	//		Once closed, close-state does not change
	CloseCh() (ch AwaitableCh)
	// IsClosed returns true is CloseCh was invoked without argument
	// and the stream was or subsequently became empty
	IsClosed() (isClosed bool)
	io.Closer
}
