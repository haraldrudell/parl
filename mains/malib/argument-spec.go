/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

const (
	// NoArguments besides switches, zero trailing arguments is allowed
	NoArguments ArgumentSpec = 1 << iota
	// OneArgument besides switches, exactly one trailing arguments is allowed
	OneArgument
	// ManyArguments besides switches, one or more trailing arguments is allowed
	ManyArguments
)

// ArgumentSpec bitfield for 0, 1, many arguments following command-line switches
//   - [NoArguments] [OneArgument] [ManyArguments]
type ArgumentSpec uint32
