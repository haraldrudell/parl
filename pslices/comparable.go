/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

type Comparable[T any] interface {
	Cmp(a, b T) (result int)
}
