/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

// go test -run=^TestOsVersion$ -v ./punix
//
//   - go version: go1.20.1
//   - os version: macOS 13.2.1
//   - or: Linux 5.15.0-56-generic
func TestOsVersion(t *testing.T) {
	fmt.Println("goversion: " + runtime.Version())
	var version string
	var hasVersion bool
	var err error
	if version, hasVersion, err = OsVersion(); err != nil {
		t.Errorf("OsVersion failed: %s", perrors.Short(err))
	} else if hasVersion {
		fmt.Println("osversion: " + version)
	}
}
