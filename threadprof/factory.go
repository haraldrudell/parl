/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package threadprof

import "time"

var HFactory HistoryFactory = &historyFactory{}

var CFactory CountersFactory = &countersFactory{}

var SFactory StatuserFactory = &statuserFactory{}

type historyFactory struct{}

func (ff *historyFactory) NewThreadHistory(useEvents bool, useHistory bool) (threadHistory History) {
	if !useEvents {
		return &threadNil{}
	}
	return newEvents(useHistory)
}

type countersFactory struct{}

func (ff *countersFactory) NewCounters(useCounters bool) (counters Counters) {
	if !useCounters {
		return &countersNil{}
	}
	return newCounters()
}

type statuserFactory struct{}

func (ff *statuserFactory) NewStatuser(useStatuser bool, d time.Duration) (statuser Statuser) {
	if !useStatuser {
		return &statuserNil{}
	}
	return newStatuser(d)
}

type threadNil struct{}

func (tn *threadNil) Event(event string, ID0 ...string) {}

func (tn *threadNil) GetEvents() (events map[string][]string) { return }

type countersNil struct{}

func (tn *countersNil) GetCounters() (list []string, m map[string]Counter) { return }
func (tn *countersNil) GetOrCreate(name string) (counter Counter)          { return &counterNil{} }
func (tn *countersNil) Reset()                                             {}

type counterNil struct{}

func (tn *counterNil) CounterValue(reset bool) (values Values) { return }
func (tn *counterNil) Dec() (value uint64)                     { return }
func (tn *counterNil) Inc() (value uint64)                     { return }
func (tn *counterNil) Value() (value uint64)                   { return }
func (tn *counterNil) Ops() (ops uint64)                       { return }
func (tn *counterNil) Max() (max uint64)                       { return }
func (tn *counterNil) IncRate() (incRate uint64)               { return }
func (tn *counterNil) DecRate() (decRate uint64)               { return }

type statuserNil struct{}

func (tn *statuserNil) Set(status string) (statuser Statuser) { return tn }
func (tn *statuserNil) Shutdown()                             {}
