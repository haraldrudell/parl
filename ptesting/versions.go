/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptesting

import (
	"runtime"

	"github.com/haraldrudell/parl/punix"
)

// Versions returns a two-line string of osversion: and goversion:
//   - "goversion: go1.20.1"
//   - "osversion: macOS 13.2.1"
//   - "osversion: Linux 5.19.0-32-generic"
//   - two benchmark file key-value pair configurations
func Versions() (twoLines string) {
	twoLines = "goversion: " + runtime.Version()

	var osVersion string
	var hasVersion bool
	var err error
	if osVersion, hasVersion, err = punix.OsVersion(); err != nil {
		panic(err)
	} else if hasVersion {
		twoLines += "\nosversion: " + osVersion
	}

	return
}
