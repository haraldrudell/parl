/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"fmt"
	"strings"
)

const (
	statusLeadin = "\x20\x20"
	leadinLen    = len(statusLeadin)
)

// GetStatus gets instance of status printer
func GetStatus() (stat *Status) {
	stat = &Status{}
	return
}

// Status handles printing of status to stdout
type Status struct {
	chars uint
}

// Print status information
func (stat *Status) Print(s string) {
	chars := len(s)
	spaceChars := int(stat.chars) - chars
	if spaceChars < 0 {
		spaceChars = 0
	}
	if chars == 0 && spaceChars == 0 {
		return
	}
	stat.chars = uint(chars)
	fmt.Printf("%s%s%s%s", statusLeadin, s,
		strings.Repeat("\x20", spaceChars),
		strings.Repeat("\b", chars+spaceChars+leadinLen))
}

// Clear remove previously printed status
func (stat *Status) Clear() {
	stat.Print("")
}
