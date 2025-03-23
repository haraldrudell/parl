/*
© 2024-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

// AwaitValue awaits value or close, blocking until either event
//   - hasValue true: value is valid, possibly the zero-value like
//     a nil interface value
//   - hasValue: false: the stream is closable and closed
//   - stream: an awaitable possibly closable source type like [Source1]
//   - — stream’s DataWaitCh Get and if present EmptyCh methods are used
//   - — stream cannot be eg. [AtomicError] because it is not awaitable
//   - AwaitValue wraps a 10-line read operation as a two-value expression
func AwaitValue[T any](stream Source1[T]) (value T, hasValue bool) {

	// endCh is a possible close channel
	//	- nil if not closable
	var endCh AwaitableCh
	if closable, isClosable := stream.(Closable[T]); isClosable {
		endCh = closable.EmptyCh()
	}

	// loop until value or closed
	for {
		select {
		case <-endCh:
			return // closable is closed return: hasValue false
		case <-stream.DataWaitCh():
			// competing with other threads for values
			//	- may receive nothing
			if value, hasValue = stream.Get(); hasValue {
				return // value read return: hasValue true, value valid
			}
		}
	}
}

// IsClosed returns true if closable is closed or triggered
//   - isClosed is a single boolean value usable with for or if
//   - IsClosed wraps a 6-line read into a single-value boolean expression
func IsClosed[T any](closable Closable[T]) (isClosed bool) {
	select {
	case <-closable.EmptyCh():
		isClosed = true
	default:
	}
	return
}
