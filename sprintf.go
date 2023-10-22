/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/plog"

// Sprintf is a printer that supports comma in large numbers
func Sprintf(format string, a ...any) string {
	if len(a) == 0 {
		return format
	}
	return plog.EnglishSprintf(format, a...)
}
