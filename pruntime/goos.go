/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

const (
	Aix       GOOS = "aix"
	Android   GOOS = "android"
	Darwin    GOOS = "darwin"
	DragonFly GOOS = "dragonfly"
	FreeBSD   GOOS = "freebsd"
	Illumos   GOOS = "illumos"
	IOS       GOOS = "ios"
	JS        GOOS = "js"
	Linux     GOOS = "linux"
	NetBSD    GOOS = "netbsd"
	OpenBSD   GOOS = "openbsd"
	Plan9     GOOS = "plan9"
	Solaris   GOOS = "solaris"
	Wasip1    GOOS = "wasm"
	Windows   GOOS = "windows"
)

// GOOS are the operating systems supported by go1.24
//   - [Aix] [Android] [Darwin] [DragonFly] [FreeBSD] [Illumos] [IOS]
//     [JS] [Linux] [NetBSD] [OpenBSD] [Plan9] [Solaris] [Wasip1] [Windows]
//   - go tool dist list
type GOOS string

func (g GOOS) String() (s string) { return string(g) }
