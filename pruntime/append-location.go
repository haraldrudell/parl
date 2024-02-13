/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"fmt"
	"strings"
)

// AppendLocation inserts [CodeLocation.Short] at the end of s, prior to a possible newline
func AppendLocation(s string, location *CodeLocation) string {
	// insert code location before a possible ending newline
	sNewline := ""
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
		sNewline = "\n"
	}
	return fmt.Sprintf("%s %s%s", s, location.Short(), sNewline)
}
