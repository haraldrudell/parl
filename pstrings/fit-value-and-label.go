/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pstrings

import "strings"

// fitValueAndLabel trums a value and its label to fit into width
//   - if pad is true and result shorter than width, s is padded with spaces to reach width
func FitValueAndLabel(width int, value, label string, pad bool) (s string) {
	if width == 0 {
		return
	}
	lengthValue := len([]rune(value))
	lengthLabel := len([]rune(label))

	// no cutting case
	if lengthValue+lengthLabel+1 <= width {
		s = value + "\x20" + label
		if pad {
			if padLength := width - (lengthValue + lengthLabel + 1); padLength > 0 {
				s += strings.Repeat("\x20", padLength)
			}
		}
		return
	}

	// value and label case: at least 1 character label
	if lengthValue+2 <= width {
		s = value + "\x20" + Fit(label, width-lengthValue-1, pad)
		return
	}

	// value alone case
	s = Fit(value, width, pad)
	return
}
