/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type Counters interface {
	GetOrCreate(name string) (counter Counter)
	GetCounters() (orderedKeys []string, m map[string]Counter)
	Reset()
}

type Counter interface {
	Inc() (value uint64)
	Dec() (value uint64)
	CounterValue(reset bool) (values CounterValues)
}

type CounterValues interface {
	Get() (value uint64, ops uint64, max uint64, incRate uint64, decRate uint64)
	Value() (value uint64)
	Ops() (ops uint64)
	Max() (max uint64)
	IncRate() (incRate uint64)
	DecRate() (decRate uint64)
}

type CountersFactory interface {
	NewCounters(useCounters bool) (counters Counters)
}
