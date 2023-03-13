/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptesting

import _ "unsafe"

// On Linux x86_64, this is aesrand seeded by /dev/urandom
//   - not thread-safe
//
//go:linkname fastrand runtime.fastrand
func fastrand() uint32
