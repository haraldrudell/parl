/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

import "flag"

type VisitedOptionsMap map[string]bool

func NewVisitedOptionsMap() (m VisitedOptionsMap) {
	m = make(map[string]bool)
	flag.Visit(m.flagVisitFunc)
	return
}

func (m *VisitedOptionsMap) flagVisitFunc(flagFlag *flag.Flag) {
	(*m)[flagFlag.Name] = true
}
