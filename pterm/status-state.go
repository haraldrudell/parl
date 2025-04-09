/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"github.com/haraldrudell/parl/ptermx"
)

// State returns state for debug purposes
func (s *StatusTerminal) State() (s2 ptermx.StatusTerminalState) {
	defer s.lock.Lock().Unlock()

	s2 = ptermx.StatusTerminalState{
		FD:               s.fd,
		IsTermTerminal:   s.isTermTerminal,
		IsTerminal:       s.isTerminal.Load(),
		Width:            s.width.Load(),
		WidthF:           s.Width(),
		StatusEnded:      s.statusEnded.Load(),
		DisplayLineCount: s.displayLineCount,
		Output:           s.output,
		CopyLog:          len(s.copyLog),
	}
	return
}
