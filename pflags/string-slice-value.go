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

// StringSliceValue manages a string-slice value for [flag.Var]
type StringSliceValue struct {
	p         *[]string
	didUpdate bool // didUpdate allows to erase the default value on first Set invocation
}

// NewStringSliceValue initializes a slice option
//   - the option’s value is stored in a slice at slicePointer
//   - defaultValue may be nil
//   - [flag.Value] is an interface of the flag package storing non-standard value types
func NewStringSliceValue(slicePointer *[]string, defaultValue []string) (v flag.Value) {
	*slicePointer = append([]string{}, defaultValue...)
	v = &StringSliceValue{p: slicePointer}
	return
}

// Set updates the string slice
//   - Set is invoked once for each option-occurrence in the command line
//   - Set appends each such option value to its list of strings
func (v *StringSliceValue) Set(optionValue string) (err error) {
	if !v.didUpdate {
		v.didUpdate = true
		*v.p = nil // clear the slice
	}
	*v.p = append(*v.p, optionValue)

	return
}

// flag package invoke String to render default value
func (v StringSliceValue) String() (s string) {
	s = strings.Join(*v.p, "\x20")
	return
}
