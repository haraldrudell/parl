/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps2

import "unsafe"

// [SwissGroupsReference] references entry allocations and points to a slice of Group
//   - each Group contains a number of entries
//   - updated go1.24 250211
//
// [SwissGroupsReference]:https://github.com/golang/go/blob/master/src/internal/runtime/maps/group.go#L293
type SwissGroupsReference struct {
	// data points to an array of groups. See groupReference above for the
	// definition of group.
	Data unsafe.Pointer // data *[length]typ.Group

	// lengthMask is the number of groups in data minus one (note that
	// length must be a power of two). This allows computing i%length
	// quickly using bitwise AND.
	LengthMask uint64
}

func (g *SwissGroupsReference) GroupSlice() {

}
