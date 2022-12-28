/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package threadprof

import (
	"time"

	"github.com/haraldrudell/parl"
)

var StatuserFactory parl.StatuserFactory = &statuserFactory{}

type statuserFactory struct{}

func (ff *statuserFactory) NewStatuser(useStatuser bool, d time.Duration) (statuser parl.Statuser) {
	if !useStatuser {
		return &statuserNil{}
	}
	return newStatuser(d)
}

type statuserNil struct{}

func (tn *statuserNil) Set(status string) (statuser parl.Statuser) { return tn }
func (tn *statuserNil) Shutdown()                                  {}

// TODO 221209 obsolete type historyFactory struct{}

// TODO 221209 obsolete type threadNil struct{}

// TODO 221209 obsolete func (tn *threadNil) Event(event string, ID0 ...parl.ThreadID) {}

// TODO 221209 obsolete func (tn *threadNil) GetEvents() (events map[parl.ThreadID][]string) { return }
