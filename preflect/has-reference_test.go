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
	const (
		pointerYes = true
		pointerNo  = false
	)

	var (
		intp         *int
		v            any
		reflectValue reflect.Value
	)

	// interface any, runtime-value nil:
	// type: <nil> value <nil> valueOf: <invalid Value> kind: invalid
	reflectValue = reflect.ValueOf(v)
	t.Logf("type: %T value %v valueOf: %s kind: %s",
		v, v,
		reflectValue.String(), reflectValue.Kind().String(),
	)

	// interface any, runtime-value *int:
	// type: *int value <nil> valueOf: <*int Value> kind: ptr
	v = intp
	reflectValue = reflect.ValueOf(v)
	t.Logf("type: %T value %v valueOf: %s kind: %s",
		v, v,
		reflectValue.String(), reflectValue.Kind().String(),
	)

	type args struct {
		v any
	}
	tests := []struct {
		name           string
		args           args
		wantHasPointer bool
	}{
		{"nil", args{nil}, pointerNo},
		{"*int", args{intp}, pointerYes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotHasPointer := HasReference(tt.args.v); gotHasPointer != tt.wantHasPointer {
				t.Errorf("HasPointer() = %v, want %v", gotHasPointer, tt.wantHasPointer)
			}
		})
	}
}
