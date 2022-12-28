/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	goidCreatorFrames = 2 // 1 for g1ID init, 1 for thread-group constructor
)

// GoEntityID is a unique named type for Go objects
type GoEntityID uint64

// GoEntityIDs generate IDs for Go Objects
var GoEntityIDs parl.UniqueIDTypedUint64[GoEntityID]

// goEntityID provides name, ID and creation time for Go objects
type goEntityID struct {
	id      GoEntityID
	t       time.Time
	creator pruntime.CodeLocation
}

type GoEntityIDer interface {
	G0ID() (id GoEntityID)
}

// newGoEntityID initiates an embedded g1ID
func newGoEntityID(extraFrames int) (g0EntityID *goEntityID) {
	if extraFrames < 0 {
		extraFrames = 0
	}
	return &goEntityID{
		id:      GoEntityIDs.ID(),
		t:       time.Now(),
		creator: *pruntime.NewCodeLocation(goidCreatorFrames + extraFrames),
	}
}

func (gi *goEntityID) G0ID() (id GoEntityID) {
	return gi.id
}
