/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package prand

import _ "unsafe"

// On Linux x86_64, this is aesrand seeded by /dev/urandom
//   - thread-safe because shared state is per-goroutine
//   - https://go.googlesource.com/go/+/go1.13.4/src/runtime/stubs.go#99
//
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

//go:linkname fastrandn runtime.fastrandn
func fastrandn(uint32) uint32
