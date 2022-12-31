/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pstrings

import (
	"strings"
)

var epsilonRune = []rune("…")[0]

// Fit return a string fitted to certain column width
//   - if width < 1, no change
//   - if s’ length equals width, no change
//   - if s’ length less than width, pad with spaces if pad true, otherwise no change
//   - otherwise cut s in the center and replace with single epsilon "…" character
func Fit(s string, width int, pad bool) (s2 string) {
	length := len([]rune(s)) // length in unicode code points

	// width < 1 or equal length or less but no pad: nothing to do
	if width < 1 || length == width || length < width && !pad {
		s2 = s
		return // no change return
	}

	if length < width {
		s2 = s + strings.Repeat("\x20", width-length)
		return
	}

	// length > width:
	// "abcdef" width:4: center:3, cut:3 right:2, left:1: "ab…ef"
	center := length / 2      // center of string, rounded down
	cut := length - width + 1 // number of characters to delete, add 1 for "…"
	right := cut/2 + 1        // characters to delete after center
	left := cut - right       // characters to delete before center
	runes := []rune(s)
	runes[center-left] = epsilonRune
	copy(runes[center-left+1:], runes[center+right:])
	runes = runes[:width]
	s2 = string(runes)
	return
}
