/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "fmt"

// StringifySlice returns the string representation of any slice
func StringifySlice[E any](slic []E) (sList []string) {
	length := len(slic)
	if length == 0 {
		return
	}
	sList = make([]string, length)
	for i, e := range slic {
		sList[i] = fmt.Sprint(e) // uses String or %v
	}
	return
}
