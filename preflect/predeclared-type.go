/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

const (
	Nil      = "nil"
	kindInt  = "int"
	kindUint = "uint"
)

// bitSize is a string containing the size of int: "(64)"
var bitSize = "(" + strconv.Itoa(int(unsafe.Sizeof(0)*8)) + ")"

// PredeclaredType returns a string describing the underlying type of value
//   - examples: "bool" "uint8" "int(64)"
//   - no named types or type aliases
func PredeclaredType(value any) (s string) {
	return reflectTypeString(reflect.TypeOf(value))
}

func reflectTypeString(typeOf reflect.Type) (s string) {
	if typeOf == nil {
		return Nil
	}
	var kindOf reflect.Kind = typeOf.Kind()
	switch kindOf {
	case reflect.Chan:
		var chanDir reflect.ChanDir = typeOf.ChanDir()
		return chanDir.String() + " " + reflectTypeString(typeOf.Elem())
	case reflect.Map:
		return "[" + reflectTypeString(typeOf.Key()) + "]" + reflectTypeString(typeOf.Elem())
	case reflect.Interface:
		return "interface " + reflectTypeString(typeOf.Elem())
	case reflect.Func:
		var sIn, sOut string
		if in := typeOf.NumIn(); in != 0 {
			sInList := make([]string, in)
			for i := range sInList {
				sInList[i] = reflectTypeString(typeOf.In(i))
			}
			sIn = "(" + strings.Join(sInList, ", ") + ")"
		} else {
			sIn = "()"
		}
		if out := typeOf.NumOut(); out != 0 {
			sOutList := make([]string, out)
			for i := range sOutList {
				sOutList[i] = reflectTypeString(typeOf.Out(i))
			}
			sOut = " (" + strings.Join(sOutList, ", ") + ")"
		}
		return kindOf.String() + sIn + sOut
	case reflect.Ptr:
		return "*" + reflectTypeString(typeOf.Elem())
	case reflect.Struct:
		var fields string
		if numFields := typeOf.NumField(); numFields > 0 {
			var sList = make([]string, numFields)
			for i := range sList {
				var structField reflect.StructField = typeOf.FieldByIndex([]int{i})
				sList[i] = reflectTypeString(structField.Type)
			}
			fields = strings.Join(sList, "; ") + " "
		}
		return "struct{ " + fields + "}"
	case reflect.Slice:
		return "[]" + reflectTypeString(typeOf.Elem())
	case reflect.Array:
		return "[" + strconv.Itoa(typeOf.Len()) + "]" + reflectTypeString(typeOf.Elem())
	case reflect.Int, reflect.Uint:
		return kindOf.String() + bitSize // "int(64)"
	default:
		return kindOf.String()
	}
}
