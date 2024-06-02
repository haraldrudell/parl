/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

//	- sink for values and slices
//	- closable sink for values and slices
//	- source for values and slices
//	- source for values and slices with all
//	- closable source for values and slices
//	- closable source for values and slices with all

// Send SendSlice SendClone
type Sink[T any] interface {
	Send(value T)
	SendSlice(values []T)
	SendClone(values []T)
}

// EmptyCh
type Closable[T any] interface {
	EmptyCh(doNotInitialize ...bool) (ch AwaitableCh)
	IsClosed() (isClosed bool)
}

type ClosableSink[T any] interface {
	// Send SendSlice SendClone
	Sink[T]
	// EmptyCh
	Closable[T]
}

// Get DataWaitCh
type Source1[T any] interface {
	Get() (value T, hasValue bool)
	DataWaitCh() (ch AwaitableCh)
	AwaitValue() (value T, hasValue bool)
}

// Get GetSlice DataWaitCh
type Source[T any] interface {
	// Get DataWaitCh
	Source1[T]
	GetSlice() (values []T)
}

// Get GetSlice GetAll DataWaitCh
type AllSource[T any] interface {
	// Get GetSlice DataWaitCh
	Source[T]
	GetAll() (values []T)
}

// Get DataWaitCh EmptyCh
type ClosableSource1[T any] interface {
	// Get DataWaitCh
	Source1[T]
	// EmptyCh
	Closable[T]
}

// Get DataWaitCh EmptyCh Init Condtion
type IterableSource[T any] interface {
	ClosableSource1[T]
	Init() (value T)
	Condition(valuep *T) (hasValue bool)
}

type ClosableSource[T any] interface {
	Source[T]
	EmptyCh(doNotInitialize ...bool) (ch AwaitableCh)
}

type ClosableAllSource[T any] interface {
	AllSource[T]
	EmptyCh(doNotInitialize ...bool) (ch AwaitableCh)
}

type SourceSink[T any] interface {
	AllSource[T]
	Sink[T]
}

type ClosableSourceSink[T any] interface {
	SourceSink[T]
	EmptyCh(doNotInitialize ...bool) (ch AwaitableCh)
}
