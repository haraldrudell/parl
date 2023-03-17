/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"strconv"

	"github.com/haraldrudell/parl"
)

// GoEntityID is a unique named type for Go objects
//   - GoEntityID is required becaue for Go objects, the thread ID is not available
//     prior to the go statement and GoGroups do not have any other unique ID
//   - GoEntityID is suitable as a map key
//   - GoEntityID uniquely identifies any Go-thread GoGroup, SubGo or SubGroup
type GoEntityID uint64

// GoEntityIDs is a generator for Go Object IDs
var GoEntityIDs parl.UniqueIDTypedUint64[GoEntityID]

// goEntityID provides name, ID and creation time for Go objects
//   - every Go object has this identifier
//   - because every Go object can also be waited upon, goEntityID also has an
//     observable wait group
//   - only public methods are G0ID() Wait() String()
type goEntityID struct {
	id GoEntityID
	wg parl.WaitGroup // Wait()
}

// newGoEntityID returns a new goEntityID that uniquely identifies a Go object
func newGoEntityID() (g0EntityID *goEntityID) {
	return &goEntityID{id: GoEntityIDs.ID()}
}

// G0ID returns GoEntityID, an internal unique idntifier
func (gi *goEntityID) G0ID() (id GoEntityID) {
	return gi.id
}

func (gi *goEntityID) Wait() { gi.wg.Wait() }

func (gi GoEntityID) String() (s string) {
	return strconv.FormatUint(uint64(gi), 10)
}
