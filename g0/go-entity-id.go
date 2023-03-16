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
type GoEntityID uint64

// GoEntityIDs generate IDs for Go Objects
var GoEntityIDs parl.UniqueIDTypedUint64[GoEntityID]

// goEntityID provides name, ID and creation time for Go objects
type goEntityID struct {
	id GoEntityID
	wg parl.WaitGroup // Wait()
}

// newGoEntityID initiates an embedded g1ID
func newGoEntityID() (g0EntityID *goEntityID) {
	return &goEntityID{id: GoEntityIDs.ID()}
}

func (gi *goEntityID) G0ID() (id GoEntityID) {
	return gi.id
}

func (gi *goEntityID) Wait() { gi.wg.Wait() }

func (gi GoEntityID) String() (s string) {
	return strconv.FormatUint(uint64(gi), 10)
}
