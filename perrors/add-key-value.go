/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
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
