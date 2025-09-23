/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

// Contains uses a map as a set returning true if key is part of the set in m
//   - key: a key of map m
//   - m a map whose keys represent set element
//   - contains: true if key is part of set m
//   - —
//   - github.com/haraldrudell/parl/pmaps.omap1.OrderedMap has Contains method
//     and convenient set-initializer MakeOrderedMapFromKeys
//     and single-value Put
func Contains[K comparable, V any](key K, m map[K]V) (contains bool) {
	_, contains = m[key]
	return
}
