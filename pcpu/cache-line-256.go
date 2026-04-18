//go:build s390x

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pcpu

const (
	// CacheLineSize is the executing processor’s
	// actual hardware cache-line size
	//	- here for s390x
	CacheLineSize = 256
)
