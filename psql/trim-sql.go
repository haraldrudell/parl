/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psql

import "strings"

// TrimSql removes leading and trailing newlines
// and replaces all infix newlines with space
func TrimSql(sql string) (trimmedSql string) {

	// trim leading and trailing newline
	trimmedSql = strings.TrimSuffix(strings.TrimPrefix(sql, "\n"), "\n")

	// remove infix newlines
	// i0 is first index not to keep to the left, i1 is last index not to keep on right
	for {
		var i0 int
		if i0 = strings.Index(trimmedSql, "\n"); i0 == -1 {
			return // no more newlines return
		}
		var i1 = i0
		// whether newline must be replaced with space separator
		var noInfix = i0 == 0
		// skip spaces left
		for i0 > 0 {
			if c := trimmedSql[i0-1]; c != '\x20' {
				noInfix = c == '(' || c == ')'
				break
			}
			i0--
			noInfix = i0 == 0
		}
		// skip spaces right
		for i1+1 < len(trimmedSql) {
			if c := trimmedSql[i1+1]; c != '\x20' {
				noInfix = c == '(' || c == ')'
				break
			}
			i1++
			noInfix = i1+1 == len(trimmedSql)
		}
		var infix string
		if noInfix {
			infix = ""
		} else {
			infix = "\x20"
		}
		trimmedSql = trimmedSql[:i0] + infix + trimmedSql[i1+1:]
	}
}
