/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

// Package ghandi interfaces Android devices
package threadprof

import (
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type CountersOn struct {
	isRunning parl.AtomicBool
	lock      sync.Mutex
	ordered   []string           // behind lock
	m         map[string]Counter // behind lock
}

func newCounters() (counters Counters) {
	return &CountersOn{m: map[string]Counter{}}
}

func (cs *CountersOn) GetOrCreate(name string) (counter Counter) {
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

func (cs *CountersOn) Add(name string) (counter Counter) {
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

func (cs *CountersOn) GetCounters() (list []string, m map[string]Counter) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	list = append([]string{}, cs.ordered...)
	m = map[string]Counter{}
	for key, value := range cs.m {
		m[key] = value
	}
	return
}
