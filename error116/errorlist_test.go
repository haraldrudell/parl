/*
© 2012–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"testing"
)

func TestErrorList(t *testing.T) {
	var err error
	var eSlice []error
	eSlice = append(eSlice, ErrorList(err)...)
	_ = eSlice

	var errorListInstance errorList
	eSlice = nil
	eSlice = append(eSlice, ErrorList(errorListInstance)...)
	_ = eSlice
}
