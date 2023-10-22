/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// EnglishSprintf is like fmt.Sprintf with thousands separator
var EnglishSprintf = message.NewPrinter(language.English).Sprintf

// sprintf is like fmt.Sprintf and:
//   - does not interpret format if a is empty, and
//   - has thousands separator for numbers
//   - is like parl.Sprintf but usable to packages imported by parl
func Sprintf(format string, a ...any) (s string) {
	if len(a) == 0 {
		return format
	}
	return EnglishSprintf(format, a...)
}
