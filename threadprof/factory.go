/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package threadprof

import (
	"time"

	"github.com/haraldrudell/parl"
)

var CountersFactory parl.CountersFactory = &countersFactory{}

var StatuserFactory parl.StatuserFactory = &statuserFactory{}

type historyFactory struct{}

type countersFactory struct{}

func (ff *countersFactory) NewCounters(useCounters bool) (counters parl.Counters) {
	if !useCounters {
		return &countersNil{}
	}
	return newCounters()
}

type statuserFactory struct{}

func (ff *statuserFactory) NewStatuser(useStatuser bool, d time.Duration) (statuser parl.Statuser) {
	if !useStatuser {
		return &statuserNil{}
	}
	return newStatuser(d)
}

type threadNil struct{}

func (tn *threadNil) Event(event string, ID0 ...parl.ThreadID) {}

func (tn *threadNil) GetEvents() (events map[parl.ThreadID][]string) { return }

type countersNil struct{}

func (tn *countersNil) GetCounters() (list []parl.CounterID, m map[parl.CounterID]parl.Counter) {
	return
}
func (tn *countersNil) GetOrCreateCounter(name parl.CounterID) (counter parl.Counter) {
	return &counterNil{}
}
func (tn *countersNil) ResetCounters() {}

type counterNil struct{}

func (tn *counterNil) CounterValue(reset bool) (values parl.CounterValues) { return }
func (tn *counterNil) Dec() (counters parl.Counter)                        { return tn }
func (tn *counterNil) Inc() (counters parl.Counter)                        { return tn }
func (tn *counterNil) Value() (value uint64)                               { return }
func (tn *counterNil) Ops() (ops uint64)                                   { return }
func (tn *counterNil) Max() (max uint64)                                   { return }
func (tn *counterNil) IncRate() (incRate uint64)                           { return }
func (tn *counterNil) DecRate() (decRate uint64)                           { return }
func (tn *counterNil) SetValue(value uint64)                               {}

type statuserNil struct{}

func (tn *statuserNil) Set(status string) (statuser parl.Statuser) { return tn }
func (tn *statuserNil) Shutdown()                                  {}
