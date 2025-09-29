/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

const (
	// do not echo to standard error on [os.Stdin] closing
	// optional argument to [Keystrokes.Launch]
	SilentClose SilentType = true
)

// optional flag for no stdin-close echo [SilentClose]
type SilentType bool
