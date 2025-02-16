/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps2

// [Group]
//   - updated go1.24 250211
//
// [Group]: https://github.com/golang/go/blob/master/src/internal/runtime/maps/group.go#L236C5-L236C9
type SwissGroup[K comparable, V any] struct {
	Ctrls CtrlGroup
	Slots [SwissMapGroupSlots]Slot[K, V]
}

// https://github.com/golang/go/blob/master/src/internal/runtime/maps/group.go#L121C1-L123C1
//   - updated go1.24 250211
type CtrlGroup uint64

// https://github.com/golang/go/blob/master/src/internal/abi/map_swiss.go#L18
//   - updated go1.24 250211
const SwissMapGroupSlots = 8

// https://github.com/golang/go/blob/master/src/internal/runtime/maps/group.go#L241C5-L245C1
//   - updated go1.24 250211
type Slot[K comparable, V any] struct {
	Key  K
	Elem V
}
