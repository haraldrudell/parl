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
	providedStringsp *[]string
	// didUpdate allows to erase the default value on first Set invocation
	//	- on first provided option, any default strings should be discarded
	didUpdate bool
}

// NewStringSliceValue initializes a slice option
//   - the option’s value is stored in a slice at slicePointer
//   - defaultValue may be nil
//   - [flag.Value] is an interface of the flag package storing non-standard value types
func NewStringSliceValue(slicePointer *[]string, defaultValue []string) (v flag.Value) {
	*slicePointer = append([]string{}, defaultValue...)
	v = &StringSliceValue{providedStringsp: slicePointer}
	return
}

// Set updates the string slice
//   - Set is invoked once for each option-occurrence in the command line
//   - Set appends each such option value to its list of strings
func (v *StringSliceValue) Set(optionValue string) (err error) {
	if !v.didUpdate {
		v.didUpdate = true
		*v.providedStringsp = nil // clear the slice
	}
	*v.providedStringsp = append(*v.providedStringsp, optionValue)

	return
}

// StringSliceValue implements flag.Value
//   - type Value interface { String() string; Set(string) error }
var _ flag.Value = &StringSliceValue{}

// flag package invoke String to render default value
func (v StringSliceValue) String() (s string) {
	if stringSlicep := v.providedStringsp; stringSlicep != nil {
		s = strings.Join(*stringSlicep, "\x20")
	}
	return
}
