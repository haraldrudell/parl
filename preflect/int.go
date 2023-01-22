/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"reflect"
)

// IntValue returns the integer value of value
//   - hasValue is true and u64 is initialized for any integer, pointer or nil
//   - for signed integer types, isSigned is true and i64 has value
//   - isPointer indicates that u64 holds a pointer value
//   - nil returns: u64 = 0, hasValue true, isPointer true, isSigned false
func IntValue(value any) (u64 uint64, i64 int64, hasValue, isSigned, isPointer bool) {
	var typeOf reflect.Type = reflect.TypeOf(value)
	if hasValue = typeOf == nil; hasValue {
		isPointer = true
		return // nil value return
	}
	var reflectValue reflect.Value = reflect.ValueOf(value)
	var reflectKind reflect.Kind = typeOf.Kind()
	switch reflectKind {
	case reflect.Ptr:
		u64 = uint64(reflectValue.Pointer())
		hasValue = true
		isPointer = true
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		u64 = reflectValue.Uint()
		hasValue = true
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		// interface conversion: interface {} is pcap.activateError, not int64
		i64 = reflectValue.Int()
		u64 = uint64(i64)
		hasValue = true
		isSigned = true
	}
	return
}
