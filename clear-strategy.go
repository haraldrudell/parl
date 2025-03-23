/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	// discards data and frees all allocations
	ClearDiscard ClearStrategy = iota + 1
	// slices freed or reallocated to minimum size
	//	- does not discard any data
	ClearScavenge
	// does nothing. Typically used to retrieve
	// allocation metrics
	ClearNoop
)

// [ClearDiscard] [ClearScavenge] [ClearNoop]
type ClearStrategy uint8
