/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptermx

import (
	"regexp"
)

const ansi = "[\u001B\u009B]" +
	"[[\\]()#;?]*(?:" +
	"(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|" +
	"(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

// TrimANSIEscapes returns a string where ANSI espace sequences
// have been removed.
func TrimANSIEscapes(s string) (s1 string) {
	return re.ReplaceAllString(s, "")
}
