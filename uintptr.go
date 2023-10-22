/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "unsafe"

// Uintptr returns v as a pointer
//   - usable with [fmt.Printf] %x
//
// Usage:
//
//	var p = &SomeStruct{}
//	parl.Log("p: 0x%x", parl.Uintptr(p))
func Uintptr(v any) (p uintptr) {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1]
}
