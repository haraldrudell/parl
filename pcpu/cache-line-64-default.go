//go:build !mips && !mipsle && !(darwin && arm64) && !ppc64 && !ppc64le && !s390x

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pcpu

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	// CacheLineSize is the executing processor’s
	// actual hardware cache-line size
	//	- here for arm64 the default and most common value
	//		across processor architectures
	CacheLineSize = 64
)

var _ = unsafe.Offsetof(struct{ x int }{}.x)
var _ = unsafe.Sizeof(CacheLineSize)
var _ unsafe.Pointer
var _ cpu.CacheLinePad
