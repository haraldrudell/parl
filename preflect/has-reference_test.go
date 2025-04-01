/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"reflect"
	"testing"
)

func TestHasReference(t *testing.T) {
	//t.Error("Logging on")

	// because HasReference is generic and
	// must be able to detect interface,
	// any-type argument cannot be used
	//	- nil cannot be provided as generic type

	var (
		hasReference bool
	)

	// byte should be false
	var b byte
	hasReference = HasReference(b)
	if hasReference {
		t.Errorf("FAIL hasReference true: %T", b)
	}

	// int* should be true
	var intp *int
	hasReference = HasReference(intp)
	if !hasReference {
		t.Errorf("FAIL hasReference false: %T", intp)
	}

	// error should be true
	var e error
	hasReference = HasReference(e)
	if !hasReference {
		// get the interface type name
		var typeName = reflect.TypeOf(&e).String()[1:]
		t.Errorf("FAIL hasReference false: %s", typeName)
	}

	// any should be true
	var a any = 3
	hasReference = HasReference(a)
	if !hasReference {
		// get the interface type name
		var typeName = reflect.TypeOf(&e).String()[1:]
		t.Errorf("FAIL hasReference false: %s", typeName)
	}

	// structs should be scanned
	var f = struct{ p any }{p: 1}
	hasReference = HasReference(f)
	if !hasReference {
		// get the interface type name
		var typeName = reflect.TypeOf(&e).String()[1:]
		t.Errorf("FAIL hasReference false: %s", typeName)
	}

	// array of int should be false
	var aInt [1]int
	hasReference = HasReference(aInt)
	if hasReference {
		t.Errorf("FAIL hasReference true: %T", aInt)
	}

	// array of pointer should be true
	var aPointer [1]*int
	hasReference = HasReference(aPointer)
	if !hasReference {
		t.Errorf("FAIL hasReference false: %T", aPointer)
	}
}
