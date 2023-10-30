/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
)

// goEntityID is an unexported unique ID for a parl.Go object
//   - every Go object has this identifier
type goEntityID struct {
	id parl.GoEntityID
}

// newGoEntityID returns a new goEntityID that uniquely identifies a Go object
func newGoEntityID() (g0EntityID *goEntityID) {
	return &goEntityID{id: parl.GoEntityIDs.ID()}
}

// EntityID returns GoEntityID, an internal unique idntifier
func (i *goEntityID) EntityID() (id parl.GoEntityID) {
	return i.id
}
