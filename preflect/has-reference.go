/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"github.com/haraldrudell/parl/preflect/prlib"
)

// HasReference returns true if v or any of its fields is of pointer type
//   - intended to detect temporary memory leaks from
//     unused elements of slices, maps and arrays
//     referring to other memory than the value itself:
//   - array slice map chan func Ptr string UnsafePointer
//
// Usage:
//
//	var v *int
//	fmt.println(preflect.HasReference) → true
func HasReference[T any](t T) (hasReference bool) {
	return prlib.HasReference(t)
}
