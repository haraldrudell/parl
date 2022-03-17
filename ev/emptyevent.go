/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

// EmptyEvent is an event with no data
type EmptyEvent struct {
	gID GoID
}

var _ Event = &EmptyEvent{}

func NewEmptyEvent(gID GoID) *EmptyEvent {
	return &EmptyEvent{gID}
}

// GoID returns the id of the sending goroutine
func (ev *EmptyEvent) GoID() (gID GoID) {
	return ev.gID
}
