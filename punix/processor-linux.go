//go:build linux

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"os"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

const (
	procCpuInfo     = "/proc/cpuinfo"
	prefixModelName = "model name\t: "
)

// Processor returns human readable string like "Intel(R) Core(TM) i5-6267U CPU @ 2.90GHz"
func Processor() (model string, err error) {

	// get byte array of cpu information
	var byts []byte
	if byts, err = os.ReadFile(procCpuInfo); perrors.Is(&err, "ReadFile %s %w", procCpuInfo, err) {
		return // read error return
	}

	// linear-search lines to find first cpu model line
	//	- there is one per core, but they are typically the same
	for _, line := range strings.Split(string(byts), "\n") {
		if m := strings.TrimPrefix(line, prefixModelName); m != line {
			model = m
			return // model found return
		}
	}

	return // model not available return
}
