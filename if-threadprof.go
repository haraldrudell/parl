/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package threadprof provide profiling of threading
package parl

import "time"

type History interface {
	Event(event string, ID0 ...string)
	GetEvents() (events map[string][]string)
}

type Counters interface {
	GetOrCreate(name string) (counter Counter)
	GetCounters() (orderedKeys []string, m map[string]Counter)
	Reset()
}

type Counter interface {
	Inc() (value uint64)
	Dec() (value uint64)
	CounterValue(reset bool) (values Values)
}

type Values interface {
	Get() (value uint64, ops uint64, max uint64, incRate uint64, decRate uint64)
	Value() (value uint64)
	Ops() (ops uint64)
	Max() (max uint64)
	IncRate() (incRate uint64)
	DecRate() (decRate uint64)
}

type HistoryFactory interface {
	NewThreadHistory(useEvents bool, useHistory bool) (history History)
}

type CountersFactory interface {
	NewCounters(useCounters bool) (counters Counters)
}

type StatuserFactory interface {
	NewStatuser(useStatuser bool, d time.Duration) (statuser Statuser)
}

type Statuser interface {
	Set(status string) (statuser Statuser)
	Shutdown()
}
