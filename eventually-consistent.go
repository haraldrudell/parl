/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

const (
	// [Awaitable.Close]: optional eventually-consistent indicator
	//   - eventual consistency increase parallel performance
	//     significantly for situations where an event must
	//     be initiated but not guaranteed to have completed
	EventuallyConsistency EventuallyConsistent = true
)

// [Awaitable.Close]: optional eventually-consistent indicator
//   - eventual consistency increase parallel performance
//     significantly for situations where an event must
//     be initiated but not guaranteed to have completed
type EventuallyConsistent bool
