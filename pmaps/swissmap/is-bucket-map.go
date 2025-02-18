/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package swissmap

import (
	"runtime"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	// runtime.Version() for go1.24
	go124 = "go1.24"
)

// IsBucketMap returns whether a legacy-style map is used,
// Go versions prior to go1.24
//   - swiss map only works with 64-bit
func IsBucketMap() (isLegacyMap bool) {
	return !pruntime.Is64Bit ||
		runtime.Version() < go124
}
