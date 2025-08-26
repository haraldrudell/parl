/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omap1

// Mapping is a type used by [OrderedMap] initializers
type Mapping[K comparable, V any] struct {
	Key   K
	Value V
}
