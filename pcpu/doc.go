/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pcpu augments the cpu package’s CPU architecture data
// eg. [CacheLineSize] the executing processor’s
// actual hardware cache-line size
package pcpu

import (
	"golang.org/x/sys/cpu"
)

var _ cpu.CacheLinePad
var value int
var _ = value

// compile-time gopls error if 32-bit architetcures have been references
// in go:build tags:
// math.MaxUint32 (untyped int constant 4294967295) overflows int [linux,mips]
// compiler (NumericOverflow)
//var _ = value > math.MaxUint32
