/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/
package cyclebreaker

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Sprintf is a printer that supports comma in large numbers
func Sprintf(format string, a ...interface{}) string {
	if len(a) == 0 {
		return format
	}
	return parlSprintf(format, a...)
}

// parlSprintf is an instantiated English-language sprintf
var parlSprintf = message.NewPrinter(language.English).Sprintf
