/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parls

import (
	"fmt"
	"strings"
)

func FilteredJoin(sList []string, separator string) (line string) {
	var nonEmpties []string
	for _, s := range sList {
		if s != "" {
			nonEmpties = append(nonEmpties, s)
		}
	}
	return strings.Join(nonEmpties, separator)
}

// QuoteList formats a string slice using %q into a single space-separated string
func QuoteList(strs []string) string {
	strs2 := make([]string, len(strs))
	for i, str := range strs {
		strs2[i] = fmt.Sprintf("%q", str)
	}
	return strings.Join(strs2, "\x20")
}

func StrSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
