/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// AnyCount is a value type that may contain
// zero, one or many values
//   - allows a single argument to represent any number of values
//   - zero or one value requires no allocation
//   - slice allocation is only required for more than one value
//
// Iteration, allocation-free:
//
//	var a AnyCount[T]
//	for i := -1; i < len(a.Values); i++ {
//	  var v T
//	  if i == -1 {
//	    if !a.HasValue {
//	      continue
//	    }
//	    v = a.Value
//	  } else {
//	    v = a.Values[i]
//	  }
//		v…
type AnyCount[T any] struct {
	// value is present when hasValue is true
	Value T
	// Values are present regardless of HasValue
	Values   []T
	HasValue bool
}

// Count returns number of values
func (a AnyCount[T]) Count() (count int) {
	count = len(a.Values)
	if a.HasValue {
		count++
	}
	return
}
