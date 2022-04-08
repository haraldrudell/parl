/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type CounterID string

type Counters interface {
	GetOrCreateCounter(name CounterID) (counter Counter)
	GetCounters() (orderedKeys []CounterID, m map[CounterID]Counter)
	ResetCounters()
}

type Counter interface {
	Inc() (counter Counter)
	Dec() (counter Counter)
	CounterValue(reset bool) (values CounterValues)
}

type CounterValues interface {
	Get() (value uint64, running uint64, max uint64, incRate uint64, decRate uint64)
	Value() (value uint64)
	Running() (running uint64)
	Max() (max uint64)
	IncRate() (incRate uint64)
	DecRate() (decRate uint64)
}

type CountersFactory interface {
	NewCounters(useCounters bool) (counters Counters)
}
