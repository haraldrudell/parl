/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type CounterID string

/*
Counters are extremely useful to determine that your code abide
by intended paralellism and identifying hangs or abnormalisms.
Printing counters every second can verify adequate progress and
possibly identify blocking of threads or swapping and garbage collection
outages.
Inc increases two counters while Dec decreases one, enabling
tracking both number of invocations as well as how many remain running by doing Inc
and a deferred Dec.
*/
type Counters interface {
	GetOrCreateCounter(name CounterID) (counter Counter)
	GetCounters() (orderedKeys []CounterID, m map[CounterID]Counter)
	ResetCounters()
}

type Counter interface {
	Inc() (counter Counter)
	Dec() (counter Counter)
	SetValue(value uint64)
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
