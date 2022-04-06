/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package threadprof

import (
	"sync"

	"github.com/haraldrudell/parl/pruntime"
)

type Events struct {
	lock       sync.Mutex
	m          map[string]*Event // behind lock
	useHistory bool
}

func newEvents(useHistory bool) (te *Events) {
	return &Events{m: map[string]*Event{}, useHistory: useHistory}
}

func (te *Events) Event(event string, ID0 ...string) {
	var ID string
	if len(ID0) > 0 {
		ID = ID0[0]
	}
	if ID == "" {
		ID = pruntime.GoRoutineID()
	}
	te.event(ID, event)
}

func (te *Events) event(ID string, event string) {
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

func (te *Events) GetEvents() (events map[string][]string) {
	events = map[string][]string{}
	te.lock.Lock()
	defer te.lock.Unlock()
	for ID, container := range te.m {
		events[ID] = container.GetEvents()
	}
	return
}
