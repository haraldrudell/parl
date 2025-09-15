/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfmt

import "fmt"

// IsPlusFlag determines if fmt.State has the '+' flag
func IsPlusFlag(s fmt.State) (is bool) { return s.Flag('+') }

// IsMinusFlag determines if fmt.State has the '-' flag
func IsMinusFlag(s fmt.State) (is bool) { return s.Flag('-') }

// IsValueVerb determines if the rune corresponds to the %v value verb
func IsValueVerb(r rune) (is bool) { return r == 'v' }

// IsStringVerb determines if the rune corresponds to the %s string verb
func IsStringVerb(r rune) (is bool) { return r == 's' }

// IsQuoteVerb determines if the rune corresponds to the %q quote verb
func IsQuoteVerb(r rune) (is bool) { return r == 'q' }
