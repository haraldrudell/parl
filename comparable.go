/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Comparable is an interface allowing for a non-comparable type
// to be ordered or compared or provide custom orderings in the slices package.
//   - Cmp(b) is expected to return an integer comparing the two parameters:
//     0 if a == b, a negative number if a < b and a positive number if a > b
//
// As of go1.19.4, Go does not support operator overloading and the comparable
// interface cannot be implemented rather only be used as a type constraint.
//
//		type T int
//		func (a T) Cmp(b T) (result int) {
//		  if a > b {
//		    return 1
//		  } else if a < b {
//		    return -1
//		  }
//		  return 0
//		}
//	 var a Comparable[T] = T(1)
//	 var b = T(2)
//	 if a.Cmp(b)…
//	 …
//		type S struct { v int }
//		func (a *S) Cmp(b *S) (result int) {
//		  if a.v > b.v {
//		    return 1
//		  } else if a.v < b.v {
//		    return -1
//		  }
//		  return 0
//		}
//	 var a Comparable[*S] = &S{1}
//	 var b = &S{2}
//	 if a.Cmp(b)…
type Comparable[T any] interface {
	Cmp(b T) (result int)
}
