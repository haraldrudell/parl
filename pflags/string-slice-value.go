/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pflags

import (
	"flag"
	"strings"
)

/*
type Value interface {
	String() string
	Set(string) error
}
*/

// StringSliceValue manages a string-slice value for flag.Var
type StringSliceValue struct {
	p         *[]string
	didUpdate bool
}

// GetStringSliceValue initializes
func GetStringSliceValue(p *[]string, value []string) (v flag.Value) {
	*p = append([]string{}, value...)
	v = &StringSliceValue{p: p}
	return
}

// Set updates the string slice
func (v *StringSliceValue) Set(s string) (err error) {
	if !v.didUpdate {
		v.didUpdate = true
		*v.p = nil
	}
	*v.p = append(*v.p, s)
	return
}

func (v StringSliceValue) String() (s string) {
	s = strings.Join(*v.p, "\x20")
	return
}
