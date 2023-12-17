/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps2

// KeyRanger is a tuple value-type
//   - used as local variable or function argument or return value causes
//     no allocation
//   - taking its address causes allocation
type KeyRanger[K comparable, V any] struct {
	List []K
	i    int
}

// RangeFunc can be used with map Range methods
func (r KeyRanger[K, V]) RangeFunc(key K, value V) (keepGoing bool) {
	r.List[r.i] = key
	r.i++
	keepGoing = r.i < len(r.List)
	return
}
