/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package ptesting provides a platform identifier for Go benchmarks.
package ptesting

import (
	"runtime"

	"github.com/haraldrudell/parl/punix"
)

const (
	goVersionKey = "goversion"
	osVersionKey = "osversion"
	cpuModelKey  = "cpu"
)

// Versions returns printable key-value pairs for the Benchmark environment:
//
// macOS:
//
//	goversion: go1.20.1
//	osversion: macOS 13.2.1
//	cpu: Apple M1 Max
//
// Linux:
//
//	goversion: go1.20.1
//	osversion: Linux 5.19.0-32-generic
//	cpu: Intel(R) Core(TM) i5-6267U CPU @ 2.90GHz
//
// Usage:
//
//	func BenchmarkXxx(b *testing.B) {
//	  println(ptesting.Versions)
//	  …
func Versions() (keyValueLines string) {

	// "goversion: go1.20.1"
	keyValueLines = goVersionKey + ":\x20" + runtime.Version()

	// "macOS 13.2.1"
	var osVersion string
	var hasVersion bool
	var err error
	if osVersion, hasVersion, err = punix.OsVersion(); err != nil {
		panic(err)
	} else if hasVersion {
		keyValueLines += "\n" + osVersionKey + ":\x20" + osVersion
	}

	// "Apple M1 Max"
	var cpuModel string
	if cpuModel, err = punix.Processor(); err != nil {
		panic(err)
	} else if cpuModel != "" {
		keyValueLines += "\n" + cpuModelKey + ":\x20" + cpuModel
	}

	return
}
