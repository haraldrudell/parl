/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "strconv"

// GoEntityID is a unique named type for Go objects
//   - GoEntityID is required becaue for Go objects, the thread ID is not available
//     prior to the go statement and GoGroups do not have any other unique ID
//   - GoEntityID is suitable as a map key
//   - GoEntityID uniquely identifies any Go-thread GoGroup, SubGo or SubGroup
type GoEntityID uint64

// GoEntityIDs is a generator for Go Object IDs
var GoEntityIDs UniqueIDTypedUint64[GoEntityID]

func (i GoEntityID) String() (s string) {
	return strconv.FormatUint(uint64(i), 10)
}
