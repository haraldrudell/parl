/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"reflect"
	"testing"
)

func TestInsertOrdered(t *testing.T) {
	type args struct {
		slice0 []int
		value  int
	}
	tests := []struct {
		name      string
		args      args
		wantSlice []int
	}{
		{"insert", args{[]int{1, 3}, 2}, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSlice := InsertOrdered(tt.args.slice0, tt.args.value); !reflect.DeepEqual(gotSlice, tt.wantSlice) {
				t.Errorf("InsertOrdered() = %v, want %v", gotSlice, tt.wantSlice)
			}
		})
	}
}

type E struct {
	value int
	ID    int
}

func Cmp(a, b E) (result int) {
	if a.value > b.value {
		return 1
	} else if a.value < b.value {
		return -1
	}
	return 0
}

func TestInsertOrderedFunc(t *testing.T) {
	type args struct {
		slice0 []E
		value  E
		cmp    func(E, E) (result int)
	}
	tests := []struct {
		name      string
		args      args
		wantSlice []E
	}{
		{"insert", args{[]E{{1, 1}, {2, 2}, {3, 3}}, E{2, 4}, Cmp}, []E{{1, 1}, {2, 2}, {2, 4}, {3, 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSlice := InsertOrderedFunc(tt.args.slice0, tt.args.value, tt.args.cmp); !reflect.DeepEqual(gotSlice, tt.wantSlice) {
				t.Errorf("InsertOrderedFunc() = %v, want %v", gotSlice, tt.wantSlice)
			}
		})
	}
}
