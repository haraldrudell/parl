/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import "testing"

func TestOffset(t *testing.T) {
	const (
		isValidNo  = false
		isValidYes = true
	)
	var (
		// sliceNil is nil slice for slice0 or slicedAway
		sliceNil []int
		// slice0 is cap 2
		slice0 = []int{1, 2}
		// slicedAway is offset1 off slice0
		slicedAway = slice0[1:]
		// slice01 is cap 1
		slice01 = []int{1}
		// slice2 has no matching slicedAway
		slice2        = []int{1, 2}
		expOffsetZero = 0
		expOffsetOne  = 1
	)
	type args struct {
		slice0     []int
		slicedAway []int
	}
	tests := []struct {
		name        string
		args        args
		wantOffset  int
		wantIsValid bool
	}{
		{"both nil", args{sliceNil, sliceNil}, 0, isValidNo},
		{"slice0 nil", args{sliceNil, slicedAway}, 0, isValidNo},
		{"slicedAway nil", args{slice0, sliceNil}, 0, isValidNo},
		{"disparate", args{slice2, slicedAway}, 0, isValidNo},
		{"slice0 cap1", args{slice01, slice01}, 0, isValidYes},
		{"offset one", args{slice0, slicedAway}, expOffsetOne, isValidYes},
		{"offset zero", args{slice0, slice0}, expOffsetZero, isValidYes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOffset, gotIsValid := Offset(tt.args.slice0, tt.args.slicedAway)
			if gotOffset != tt.wantOffset {
				t.Errorf("Offset() gotOffset = %v, want %v", gotOffset, tt.wantOffset)
			}
			if gotIsValid != tt.wantIsValid {
				t.Errorf("Offset() gotIsValid = %v, want %v", gotIsValid, tt.wantIsValid)
			}
		})
	}
}
