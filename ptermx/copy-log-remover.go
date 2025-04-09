/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptermx

const (
	// [StatusTerminal.CopyLog] remove writer
	CopyLogRemove CopyLogRemover = iota + 1
)

// argument type for [StatusTerminal.CopyLog]
// - [CopyLogRemove]
type CopyLogRemover uint8
