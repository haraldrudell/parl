/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package prlib

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

	// because T is generic type, t cannot be untyped nil
	//	- for concrete types, [reflect.TypeOf] and [reflect.ValueOf] work
	//	- for interface types, those methods reflect the dynamic type,
	//		ie. the currently assigned value
	//	- if dynamic type is nil, those methods report for nil
	//	- to determine whether T is interface, use go1.22 [reflect.TypeFor]
	var reflectType = reflect.TypeFor[T]()
	return traverseForReference(reflectType)
}

// traverseForReference returns true if reflectType or any of
// its possible struct fields may hold a reference
func traverseForReference(reflectType reflect.Type) (hasReference bool) {

	// fmt.Printf("type: %s kind: %s\n", reflectType, reflectType.Kind())

	// kind that reflects interface types
	var kind = reflectType.Kind()

	// array: determined by its element type
	if kind == reflect.Array {
		kind = reflectType.Elem().Kind()
	}

	hasReference = KindHasReference(kind)

	if hasReference || kind != reflect.Struct {
		return
	}
	// kind is struct

	// traverse struct fields
	for i := range reflectType.NumField() {
		var field = reflectType.Field(i)
		if hasReference = traverseForReference(field.Type); hasReference {
			return
		}
	}

	return // struct no references
}

// KindHasReference returns true if kind may hold references
//   - kind [reflect.Struct]: returned as false, but fields needs to be iterated
//   - kind [reflect.Array]: returned as false, but element type needs to be examined
func KindHasReference(kind reflect.Kind) (hasReference bool) {

	switch kind {
	case reflect.Ptr, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map,
		reflect.Slice, reflect.String,
		reflect.UnsafePointer:
		return true
	case reflect.Array:

	case reflect.Struct:
	}
	// kind does not have reference

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
