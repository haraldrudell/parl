/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

// DataEvent is an event with an empty interface data value.
// The goroutine defines an exported unique type that it sends using ctx.Send(payload).
// The main evvent loop checks the type of the pyload against the exported unique type
type DataEvent struct {
	EmptyEvent
	payload interface{}
}

var _ Event = &DataEvent{}

// NewDataEvent creates an event with payload of a single data value
func NewDataEvent(gID GoID, payload interface{}) (evt Event) {
	return &DataEvent{*NewEmptyEvent(gID), payload}
}

// Payload returns the data value of a DataEvent
func (ev *DataEvent) Payload() (payload interface{}) {
	return ev.payload
}
