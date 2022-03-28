/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parlp provides portable computer process information
package parlp

import (
	"time"

	gosysinfo "github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"github.com/haraldrudell/parl"
)

// ProcessStartTime returns the time the executing process was started.
// The package used is:
//   import "github.com/elastic/go-sysinfo"
// This packlage was found on 220322 using Go package search: https://pkg.go.dev/search?q=sysinfo
func ProcessStartTime() (t time.Time) {

	var process types.Process
	var err error
	if process, err = gosysinfo.Self(); err != nil {
		panic(parl.Errorf("go-sysinfo.Self: %w", err))
	}

	var processInfo types.ProcessInfo
	if processInfo, err = process.Info(); err != nil {
		panic(parl.Errorf("go-sysinfo.Info: %w", err))
	}

	return processInfo.StartTime
}
