/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "unsafe"

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
