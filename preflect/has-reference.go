/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"reflect"
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

	// a representation of the value, reflect.Value
	var reflectValue = reflect.ValueOf(t)
	// a representation of the type
	var kind = reflectValue.Kind()

	hasReference = kindHasReference(kind)

	//parl.D("kind: %s has: %t", kind.String(), hasReference)

	return
}

func kindHasReference(kind reflect.Kind) (hasReference bool) {
	switch kind {
	case reflect.Ptr, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map,
		reflect.Slice, reflect.Uintptr, reflect.String,
		reflect.UnsafePointer:
		return true
	case reflect.Struct:
	}

	return
}

// Kinds:
// Bool: a boolean value
// Int, Int8, Int16, Int32, Int64: an integer value
// Uint, Uint8, Uint16, Uint32, Uint64, Uintptr: an unsigned integer value
// Float32, Float64: a floating-point value
// Complex64, Complex128: a complex number
// Array: an array
// Chan: a channel
// Func: a function
// Interface: an interface
// Map: a map
// Ptr: a pointer
// Slice: a slice
// String: a string
// Struct: a struct
// UnsafePointer: an unsafe pointer
