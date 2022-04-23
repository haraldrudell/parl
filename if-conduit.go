/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type Sender[T any] interface {
	// Put is thread-safe, non-blocking, panic-free.
	Put(value T) (IsCanceled bool)
	// PutSlice is thread-safe, non-blocking, panic-free.
	// PutSlice(nil) determines if the counduit is canceled.
	PutSlice(values []T) (IsCanceled bool)
}

type Receiver[T any] interface {
	// Get receives one element from the conduit.
	// ok is true if Get did receive a valid element.
	// Get blocks if no element is available.
	// Get returns the zero-value of T and ok false if the coundit
	// is closed and empty.
	Get() (value T, ok bool)
	// GetSlice receives one or more elements from the conduit.
	// ok is true if GetSlice did receive valid elements.
	// GetSlice blocks if no elements are available.
	// GetSlice returns the zero-value of T and ok false if the coundit
	// is closed and empty.
	GetSlice(max int) (values []T, ok bool)
	// IsCanceled determines if the conduit has been canceled.
	// The conudit may still have data elements.
	IsCanceled() (IsCanceled bool)
	// IsEmpty determines if the conduit is empty
	IsEmpty() (isEmpty bool)
	// Count retrieves how many data elements are waiting in the counduit
	Count() (count int)
	// WaitCount retrieves how many conduit receive invocations are waiting for data
	WaitCount() (waitCount int)
}

type Conduit[T any] interface {
	Sender[T]
	Receiver[T]
}
