/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"math"
	"testing"
	"unsafe"
)

func TestIntValue(t *testing.T) {
	hasValueTrue := true
	hasValueFalse := false
	isSignedFalse := false
	isSignedTrue := true
	isPointerTrue := true
	isPointerFalse := false
	type args struct {
		value any
	}
	tests := []struct {
		name          string
		args          args
		wantU64       uint64
		wantI64       int64
		wantHasValue  bool
		wantIsSigned  bool
		wantIsPointer bool
	}{
		{"nil", args{nil}, 0, 0, hasValueTrue, isSignedFalse, isPointerTrue},
		{"uint", args{uint(1)}, 1, 0, hasValueTrue, isSignedFalse, isPointerFalse},
		{"int", args{-1}, math.MaxUint64, -1, hasValueTrue, isSignedTrue, isPointerFalse},
		{"non-integer", args{false}, 0, 0, hasValueFalse, isSignedFalse, isPointerFalse},
		{"pointer", args{&hasValueTrue}, uint64(uintptr(unsafe.Pointer(&hasValueTrue))), 0, hasValueTrue, isSignedFalse, isPointerTrue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotU64, gotI64, gotHasValue, gotIsSigned, gotIsPointer := IntValue(tt.args.value)
			if gotU64 != tt.wantU64 {
				t.Errorf("IntValue() gotU64 = %v, want %v", gotU64, tt.wantU64)
			}
			if gotI64 != tt.wantI64 {
				t.Errorf("IntValue() gotI64 = %v, want %v", gotI64, tt.wantI64)
			}
			if gotHasValue != tt.wantHasValue {
				t.Errorf("IntValue() gotHasValue = %v, want %v", gotHasValue, tt.wantHasValue)
			}
			if gotIsSigned != tt.wantIsSigned {
				t.Errorf("IntValue() gotIsSigned = %v, want %v", gotIsSigned, tt.wantIsSigned)
			}
			if gotIsPointer != tt.wantIsPointer {
				t.Errorf("IntValue() gotIsPointer = %v, want %v", gotIsPointer, tt.wantIsPointer)
			}
		})
	}
}
