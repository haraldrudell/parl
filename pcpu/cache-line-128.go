//go:build (darwin && arm64) || ppc64 || ppc64le

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pcpu

const (
	// CacheLineSize is the executing processor’s
	// actual hardware cache-line size
	//	- here for darwin-arm64, ppc64 and pp64le
	CacheLineSize = 128
)
