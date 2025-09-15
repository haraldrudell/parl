/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// ChainStringer obtains a comprehensive string representation of an error chain.
// formats used are DefaultFormat ShortFormat LongFormat ShortSuffix LongSuffix
type ChainStringer interface {
	// ChainString is used by ChainStringer() to obtain comprehensive
	// string representations of errors.
	// The argument isIgnore() is used to avoid printing cyclic error values.
	// If a ChainStringer pointer receiver gets a nil value, the empty string is returned.
	// ChainString() obtains a string representation of the errors in its chain.
	// Rich errors implement either ChainStringer or fmt.Formatter
	ChainString(format CSFormat) string
}
