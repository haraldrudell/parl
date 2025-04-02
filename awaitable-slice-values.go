/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type AwaitableSliceValues[T any] struct {
	Head    []T
	Outputs [][]T
	Primary []T
	Inputs  [][]T
}

const (
	// [AwaitableSlice.State] populate slice values
	ValuesYes ValuesFlag = iota + 1
)

// [AwaitableSlice.State] ValuesYes: populate slice values
type ValuesFlag uint8

// getValues copies elements from the queue to allocated slices
//   - both outQ and InQ locks must be held
func (s *AwaitableSlice[T]) getValues(values *AwaitableSliceValues[T]) {
	if x := len(s.outQ.head); x > 0 {
		values.Head = make([]T, x)
		copy(values.Head, s.outQ.head)
	}
	if x := len(s.outQ.sliceList); x > 0 {
		values.Outputs = make([][]T, x)
		for i, src := range s.outQ.sliceList {
			var dest = &values.Outputs[i]
			*dest = make([]T, len(src))
			copy(*dest, src)
		}
	}

	if x := len(s.outQ.InQ.primary); x > 0 {
		values.Primary = make([]T, x)
		copy(values.Primary, s.outQ.InQ.primary)
	}
	if x := len(s.outQ.InQ.sliceList); x > 0 {
		values.Inputs = make([][]T, x)
		for i, src := range s.outQ.InQ.sliceList {
			var dest = &values.Inputs[i]
			*dest = make([]T, len(src))
			copy(*dest, src)
		}
	}
}
