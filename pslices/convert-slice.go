/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ConvertSliceToInterface converts a slice of a struct type to a slice of an interface type.
package pslices

import "github.com/haraldrudell/parl/perrors"

// ConvertSliceToInterface converts a slice of a struct type to a slice of an interface type.
//
//   - T is any type
//   - G is an interface type that *T implements
//
// If *T does not implement G: runtime panic
func ConvertSliceToInterface[T, G any](structSlice []T) (interfaceSlice []G) {
	length := len(structSlice)
	interfaceSlice = make([]G, length)
	for i := 0; i < length; i++ {
		valueAny := any(&structSlice[i])
		var ok bool
		if interfaceSlice[i], ok = valueAny.(G); !ok {
			var t T
			var g G
			panic(perrors.ErrorfPF("type assertion failed: %T does not implement %T", t, g))
		}
	}
	return
}
