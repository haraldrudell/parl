/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

import "github.com/google/uuid"

// EvThread holds manager information for a running goroutine
type EvThread struct {
	ID   GoID
	Name string
}

// NewEvThread holds manager information for a running goroutine
func NewEvThread(name string) (ti *EvThread) {
	return &EvThread{ID: GoID(uuid.New()), Name: name}
}
