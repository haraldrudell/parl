/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

// ExitEvent holds the outcome of a goruotine
type ExitEvent struct {
	EmptyEvent
	Err       error
	IsCancel  bool
	IsTimeout bool
}

// NewExitEvent holds the outcome of a goruotine
func NewExitEvent(e error, gID GoID) (evt Event) {
	return &ExitEvent{EmptyEvent: *NewEmptyEvent(gID), Err: e}
}
