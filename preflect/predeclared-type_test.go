/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package preflect

import (
	"strconv"
	"testing"
	"unsafe"
)

func TestPredeclaredType(t *testing.T) {
	var bitSize = int(unsafe.Sizeof(0) * 8)
	type args struct {
		value any
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		{"nil", args{nil}, "nil"},
		{"bool", args{false}, "bool"},
		{"uint8", args{uint8(0)}, "uint8"},
		{"uint", args{uint(0)}, "uint(" + strconv.Itoa(bitSize) + ")"},
		{"byte", args{byte(0)}, "uint8"},
		{"complex64", args{complex64(0)}, "complex64"},
		{"float32", args{float32(0)}, "float32"},
		{"uintptr", args{uintptr(0)}, "uintptr"},
		{"string", args{""}, "string"},
		{"array", args{[1]int{0}}, "[1]int(64)"},
		{"slice", args{[]int{0}}, "[]int(64)"},
		{"struct", args{struct{ int }{}}, "struct{ int(64) }"},
		{"pointer", args{&bitSize}, "*int(64)"},
		{"function", args{func(i int) (s string) { return }}, "func(int(64)) (string)"},
		{"interface nil", args{error(nil)}, "nil"},
		{"interface non-nil", args{any(1)}, "int(64)"},
		{"function", args{func(i int) (s string) { return }}, "func(int(64)) (string)"},
		{"map", args{map[int]string{}}, "[int(64)]string"},
		{"chan", args{make(chan int)}, "chan int(64)"},
		{"chan", args{make(chan<- bool)}, "chan<- bool"},
		{"chan", args{make(<-chan struct{})}, "<-chan struct{ }"},
		{"typed nil", args{(*int)(nil)}, "*int(64)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := PredeclaredType(tt.args.value); gotS != tt.wantS {
				t.Errorf("PredeclaredType() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
