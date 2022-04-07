/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type History interface {
	Event(event string, ID0 ...string)
	GetEvents() (events map[string][]string)
}

type HistoryFactory interface {
	NewThreadHistory(useEvents bool, useHistory bool) (history History)
}
