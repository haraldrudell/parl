/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

const (
	// [PtrSize] is the size of a pointer in bytes - unsafe.Sizeof(uintptr(0)) but as an ideal constant.
	// It is also the size of the machine's native word size (that is, 4 on 32-bit systems, 8 on 64-bit).
	//	-
	//	- ^uintptr(0): a pointer of bits all set to 1
	//	- ^uintptr(0) >> 63: 0 on 32-bit, 1 on 64-bit
	//	- PtrSize: 4 or 8
	//
	// [PtrSize]: https://github.com/golang/go/blob/master/src/internal/goarch/goarch.go#L31C1-L34C1
	PtrSize = 4 << (^uintptr(0) >> 63)
	// Is32Bit is true if prpocessor architure is 32-bit
	Is32Bit = ^uintptr(0)>>63 == 0
	// Is64Bit is true if prpocessor architure is 64-bit
	Is64Bit = ^uintptr(0)>>63 == 1
)
