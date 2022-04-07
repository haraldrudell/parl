/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package threadprof

import (
	"sync"

	"github.com/haraldrudell/parl/goid"
)

type Events struct {
	lock       sync.Mutex
	m          map[goid.ThreadID]*Event // behind lock
	useHistory bool
}

func newEvents(useHistory bool) (te *Events) {
	return &Events{m: map[goid.ThreadID]*Event{}, useHistory: useHistory}
}

func (te *Events) Event(event string, ID0 ...goid.ThreadID) {
	var ID goid.ThreadID
	if len(ID0) > 0 {
		ID = ID0[0]
	}
	if ID == "" {
		ID = goid.GoRoutineID()
	}
	te.event(ID, event)
}

func (te *Events) event(ID goid.ThreadID, event string) {
	te.lock.Lock()
	defer te.lock.Unlock()
	var ec *Event
	var ok bool
	if ec, ok = te.m[ID]; !ok {
		ec = newEventContainer(te.useHistory)
		te.m[ID] = ec
	}
	ec.Event(event)
}

func (te *Events) GetEvents() (events map[goid.ThreadID][]string) {
	events = map[goid.ThreadID][]string{}
	te.lock.Lock()
	defer te.lock.Unlock()
	for ID, container := range te.m {
		events[ID] = container.GetEvents()
	}
	return
}
