/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/goid"

// History may be deprecated in favor of Tracer
type History interface {
	Event(event string, ID0 ...goid.ThreadID)
	GetEvents() (events map[goid.ThreadID][]string)
}

type HistoryFactory interface {
	NewThreadHistory(useEvents bool, useHistory bool) (history History)
}
