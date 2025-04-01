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
