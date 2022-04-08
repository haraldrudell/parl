/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ghandi interfaces Android devices
package threadprof

import (
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type Counters struct {
	isRunning parl.AtomicBool
	lock      sync.Mutex
	ordered   []parl.CounterID                // behind lock
	m         map[parl.CounterID]parl.Counter // behind lock
}

func newCounters() (counters parl.Counters) {
	return &Counters{m: map[parl.CounterID]parl.Counter{}}
}

func (cs *Counters) GetOrCreateCounter(name parl.CounterID) (counter parl.Counter) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	var ok bool
	if counter, ok = cs.m[name]; ok {
		return
	}
	counter = newCounter()
	cs.ordered = append(cs.ordered, name)
	cs.m[name] = counter
	return
}

func (cs *Counters) Add(name parl.CounterID) (counter parl.Counter) {
	if cs.isRunning.IsTrue() {
		panic(perrors.Errorf("Add while Counter running: %s", name))
	}
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if _, ok := cs.m[name]; ok {
		panic(perrors.Errorf("Counter already exist: %s", name))
	}
	counter = newCounter()
	cs.ordered = append(cs.ordered, name)
	cs.m[name] = counter
	return
}

func (cs *Counters) GetCounters() (list []parl.CounterID, m map[parl.CounterID]parl.Counter) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	list = append([]parl.CounterID{}, cs.ordered...)
	m = map[parl.CounterID]parl.Counter{}
	for key, value := range cs.m {
		m[key] = value
	}
	return
}

func (cs *Counters) ResetCounters() {
	_, m := cs.GetCounters()
	for _, counter := range m {
		counter.CounterValue(true)
	}
}
