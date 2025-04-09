/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptermx

import "fmt"

type StatusTerminalState struct {
	FD                         int
	IsTermTerminal, IsTerminal bool
	Width, WidthF              int
	StatusEnded                bool
	DisplayLineCount           int
	Output                     string
	CopyLog                    int
}

func (s *StatusTerminalState) String() (s2 string) {
	return fmt.Sprintf(
		"fd %d isTT %t isT %t W %d %d ended %t lines %d copies %d status: %d%q",
		s.FD, s.IsTermTerminal, s.IsTerminal, s.Width, s.WidthF,
		s.StatusEnded, s.DisplayLineCount, s.CopyLog,
		len(s.Output), s.Output,
	)
}
