/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"slices"

	"github.com/haraldrudell/parl/perrors/errorglue"
)

// AddKeyValue attaches a string value to err
//   - values can be retrieved using [ErrorData]
//   - if key is non-empty valiue is returned in a map where last key wins
//   - if key is empty, valuse is returned in s string slice
//   - err can be nil
func AddKeyValue(err error, key, value string) (e error) {
	return errorglue.NewErrorData(err, key, value)
}

// ErrorData returns any embedded data values from err and its error chain as a list and map
//   - list contains values where key was empty, oldest first
//   - keyValues are string values associated with a key string, overwriting older values
//   - err list keyValues may be nil
func ErrorData(err error) (list []string, keyValues map[string]string) {

	// traverse the err and its error chain, newest first
	//	- only errors with data matter
	for ; err != nil; err = errors.Unwrap(err) {

		// ignore errrors without key/value pair
		var e, ok = err.(errorglue.ErrorHasData)
		if !ok {
			continue
		}

		// empty key is appended to slice
		//	- oldest value first
		var key, value = e.KeyValue()
		if key == "" { // for the slice
			list = append(list, value) // newest first
			continue
		}

		// for the map
		if keyValues == nil {
			keyValues = map[string]string{key: value}
			continue
		}
		// values are added newset first
		//	- do not overwrite newer values with older
		if _, ok := keyValues[key]; !ok {
			keyValues[key] = value
		}
	}
	slices.Reverse(list) // oldest first

	return
}
