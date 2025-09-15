/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"strings"
)

// DumpChain retrieves a space-separated string of
// error implementation type-names found in the main error
// chain of err
//   - err: the error to traverse.
//     err can be nil
//   - typeNames: list of type-names
//
// Usage:
//
//	fmt.Println(perrors.Stack(errors.New("an error")))
//	*error116.errorStack *errors.errorString
func DumpChain(err error) (typeNames string) {
	var strs []string
	for err != nil {
		strs = append(strs, fmt.Sprintf("%T", err))
		err, _, _ = Unwrap(err)
	}
	typeNames = strings.Join(strs, "\x20")
	return
}
