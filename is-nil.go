/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "unsafe"

// IsNil checks whether an interface value is truly nil
//   - In Go, comparison of an interface value that has been assigned
//     a concretely typed nil value yields unexpected results
//   - (any)((*int)(nil)) == nil → false, where true is expected
//   - IsNil((*int)(nil)) → true
//   - as of go1.20.3, an interface value is 2 pointers,
//   - — the first currently assigned type and
//   - —the second currently assigned value
func IsNil(v any) (isNil bool) {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
	/*
		if isNil = v == nil; isNil {
			return // untyped nil: true return
		}
		reflectValue := reflect.ValueOf(v)
		isNil = reflectValue.Kind() == reflect.Ptr &&
			reflectValue.IsNil()
		return
	*/
}

func Uintptr(v any) (p uintptr) {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1]
}
