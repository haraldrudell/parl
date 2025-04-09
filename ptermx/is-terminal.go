/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptermx

const (
	// [StatusTerminal.SetTerminal] default isTerminal value
	NoIsTerminal IsTerminal = iota + 1
	// [StatusTerminal.SetTerminal] activate isTerminal override
	IsTerminalYes
)

// determines status output override for [StatusTerminal.SetTerminal]
//   - [IsTerminalYes] [NoIsTerminal]
type IsTerminal uint8
