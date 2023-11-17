/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import _ "unsafe"

// CpuTick is a 32-bit pseudo-random number increasing every ns
//   - invocation time 16 ns, similar to an uncontended sync.Mutex.Lock
//   - wraps around every 4 s
//   - getting a random number is 4 ns
//   - — Intn on single-threaded math/rand.New(rand.NewSource(time.Now().UnixNano()))
//
// On Linux x86_64, this is aesrand seeded by /dev/urandom
//   - thread-safe because shared state is per-goroutine
//   - https://go.googlesource.com/go/+/refs/heads/master/src/runtime/stubs.go#124
//
//go:linkname CpuTicks runtime.cputicks
func CpuTicks() uint32
