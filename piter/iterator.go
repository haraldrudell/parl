/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package piter

import "iter"

// Iterator is an for range iterator over T
type Iterator[T any] interface {
	// Seq is an iterator over sequences of individual values.
	// When called as seq(yield), seq calls yield(v) for
	// each value v in the sequence, stopping early if yield returns false.
	Seq(yield func(value T) (keepGoing bool))
}

// Iterator.Seq is iter.Seq
//   - type Seq[V any] func(yield func(V) bool)
var _ = func(i Iterator[int]) (seq iter.Seq[int]) { return i.Seq }
