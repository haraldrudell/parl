/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pstrings

import (
	"fmt"
	"strings"
)

// FilteredJoinWithHeading takes a string slice of heading-value pairs and
// only outputs those strings that are non-empty.
// {"head1", "first", "head2, ""} → "head1: first"
func FilteredJoinWithHeading(sList []string, separator ...string) (line string) {
	var sep string
	if len(separator) > 0 {
		sep = separator[0]
	} else {
		sep = "\x20"
	}
	var nonEmpties []string
	for i := 0; i < len(sList)-1; i += 2 {
		heading := sList[i]
		value := sList[i+1]
		if len(value) == 0 {
			continue
		}
		if len(heading) > 0 {
			value = heading + ": " + value
		}
		nonEmpties = append(nonEmpties, value)
	}
	return strings.Join(nonEmpties, sep)
}

// FilteredJoin is like strings.Join but ignores empty strings.
// defauklt separator is single space \x20
func FilteredJoin(sList []string, separator ...string) (line string) {
	var sep string
	if len(separator) > 0 {
		sep = separator[0]
	} else {
		sep = "\x20"
	}
	var nonEmpties []string
	for _, s := range sList {
		if s != "" {
			nonEmpties = append(nonEmpties, s)
		}
	}
	return strings.Join(nonEmpties, sep)
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
