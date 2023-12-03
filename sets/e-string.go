/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"
	"reflect"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// eString is an improved %v string converter.
// E is a concrete type, that may be indirect, or interface
//   - E can be:
//   - — a type without String method
//   - — a type with pointer-receiver String method
//   - — a type with value-receiver String method
//   - — a pointer to those
//   - issues:
//   - — if E is pointer to type without String method,
//     a uintptr pointer is output ‘0x1400000e4b0’ where
//     it is desirable to print values not pointers
//   - — if E has pointer-receiver String but is not pointer,
//     type assertion fails to find the String method
//   - —
//   - if e is basic type, %v: eString(1) → ‘1’
//   - if e is *basic type, %v: eString(&1) → ‘1’
//   - if e is *basic type, %v: eString(&abc) → ‘abc’
//   - if e is String pointer receiver: eString((*) String()) → ‘string’
//   - if e is *String pointer receiver: eString(&(*) String()) → ‘string’
//   - if e is String value receiver: eString(() String()) → ‘string’
//   - if e is *String value receiver: eString(&() String()) → ‘string’
func eString[E any](e E) (s string) {

	// this type assertion works if:
	//	- e is any(struct) where struct has value-receiver String method
	//	- e is any(&struct) where struct has value-receiver String method
	//	- e is any(&struct) where struct has pointer-receiver String method
	// this type assertion fails if:
	//	- e is nil
	//	- e does not have String method
	//	- e has more than one indirection
	//	- e is non-pointer struct with pointer-receiver String method
	if stringer, ok := any(e).(fmt.Stringer); ok {
		s = stringer.String()
		return
	}

	// handle case where:
	//	- e is non-pointer struct with pointer-receiver String method
	//	- because e is concrete type, we can use address operator
	//	- and then convert that to interface and do type assertion
	if stringer, ok := any(&e).(fmt.Stringer); ok {
		s = stringer.String()
		return
	}

	// handle non-pointer types and nil
	var isNil = cyclebreaker.IsNil(e)
	var typeName string
	var isPointer bool
	if !isNil {
		typeName = fmt.Sprintf("%T", e)
		isPointer = typeName[0] == '*'
	}
	if isNil || !isPointer {
		// some basic or composite type value or
		//	- typed or untyped nil pointer: ‘<nil>’
		s = fmt.Sprintf("%v", e)
		return
	}

	// E is non-nil pointer to something
	//	- it is not allowed to indirect E and
	//	- it is not desirable to print using default %v: uintptr like ‘0x1400000e4c8’
	//	- if the value is extracted using unsafe.Pointer,
	//		%v will print whatever value the unsafe.Pointer is claimed to be
	//	- since E is a real type and not any, reflection will work
	//	- reflection can extract the pointed-to value and assign it to
	//		an interface variable

	// e is not nil and known to be pointer
	//	- the concrete value stored in e
	var reflectValue = reflect.ValueOf(e)
	// extract the value E is pointing to
	var pointedToValue = reflectValue.Elem()
	// store the pointed-to value as interface value
	var anyValue = pointedToValue.Interface()
	// print anyValue’s dynamic type using %v
	s = fmt.Sprintf("%v", anyValue)

	return
}
