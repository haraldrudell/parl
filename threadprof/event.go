/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package threadprof

type Event struct {
	v interface{}
}

func newEventContainer(useHistory bool) (ev *Event) {
	e := Event{}
	if !useHistory {
		e.v = ""
	}
	return &e
}

func (ev *Event) Event(event string) {
	if _, ok := ev.v.(string); ok {
		ev.v = event
		return
	}
	if ev.v == nil {
		ev.v = []string{event}
		return
	}
	list := ev.v.([]string)
	ev.v = append(list, event)
}

func (ev *Event) GetEvents() (events []string) {
	if ev.v == nil {
		return
	}

	if event, ok := ev.v.(string); ok {
		if event != "" {
			events = []string{event}
		}
		return
	}

	list := ev.v.([]string)
	events = make([]string, len(list))
	copy(events, list)
	return
}
