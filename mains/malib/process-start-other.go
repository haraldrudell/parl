//go:build !darwin

/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// format absolute path into /proc directory
	filePrintf           = "/proc/%d/stat"
	rightParenthesisByte = byte(')')
	// the number of clock ticks per second
	//	- getconf CLK_TCK
	//	- 100
	//	- man sysconf, clock ticks
	//	- The  number  of  clock  ticks  per second
	clockTicksPerSecond = 100
	clockTickNs         = time.Second / clockTicksPerSecond
	// starttime index, 1-based, in /proc/[pid]/stat space-separated file.
	//	- man proc, /proc/[pid]/stat, entry (22) 1-based
	//	- starttime  %llu
	//	- The time the process started after system boot
	//	- the value is expressed in clock  ticks  (divide  by sysconf(_SC_CLK_TCK)).
	startTimeIndex = 21 - 2
	// path of /proc/uptime
	//	- This file contains two numbers
	//	- man proc
	//	- values in seconds: the uptime of the system
	//	- cat /proc/uptime
	// 	- 5422217.66 21636302.50
	procUptime      = "/proc/uptime"
	procUptimeField = 0
)

// ProcessStartTime returns start time for process pid with second resolution
//   - panic on troubles
func ProcessStartTime() (createTime time.Time) {
	var err error
	if createTime, err = ProcessStart(os.Getpid()); err != nil {
		panic(err)
	}
	return
}

// ProcessStart returns start time for process pid with second resolution
func ProcessStart(pid int) (processStart time.Time, err error) {

	// get system uptime resolution 10 ms
	var uptime time.Duration
	if uptime, err = systemUptime(); err != nil {
		return
	}

	// get process start time in clock ticks
	var startTicks int64
	if startTicks, err = processStartTicks(pid); err != nil {
		return
	}

	// get process duration ± 20 ms
	var processDuration = uptime - time.Duration(startTicks)*clockTickNs

	// calculate process start time
	processStart = time.Now().Add(-processDuration).Truncate(time.Second)

	return
}

// SystemUpSince returns boot time second resolution
func SystemUpSince() (upSince time.Time, err error) {
	var uptime time.Duration
	if uptime, err = systemUptime(); err != nil {
		return
	}
	upSince = time.Now().Add(-uptime).Truncate(time.Second)
	return
}

// systemUptime returns host up time resolution 10 ms
func systemUptime() (uptime time.Duration, err error) {

	// read /proc/uptime
	var data []byte
	if data, err = os.ReadFile(procUptime); perrors.Is(&err, "os.ReadFile %w", err) {
		return
	}

	// extract numeric string
	var fields = bytes.Fields(data)
	if procUptimeField >= len(fields) {
		err = perrors.ErrorfPF("content too short: %q", procUptime)
		return
	}
	// remove the period
	var uptimeS = strings.Replace(string(fields[procUptimeField]), ".", "", 1)

	// convert from unit 10 ms to time.Duration
	var u64 uint64
	if u64, err = strconv.ParseUint(uptimeS, 10, 64); perrors.Is(&err, "ParseUint %w", err) {
		return
	}
	uptime = time.Duration(u64) * 10 * time.Millisecond

	return
}

// processStartTicks returns the time in system clock ticks since boot when the process was started
func processStartTicks(pid int) (clockTicks int64, err error) {

	// read /proc/n/stat
	var data []byte
	var filename = fmt.Sprintf(filePrintf, pid)
	if data, err = os.ReadFile(filename); perrors.Is(&err, "os.ReadFile %w", err) {
		return
	}

	// get tick count
	var index int
	if index = bytes.LastIndexByte(data, rightParenthesisByte); index == -1 {
		err = perrors.ErrorfPF("no %q found in %q", rightParenthesisByte, filename)
		return
	}
	var fields = bytes.Fields(data[index+1:])
	if startTimeIndex >= len(fields) {
		err = perrors.ErrorfPF("content too short: %q", filename)
		return
	}
	clockTicks, err = strconv.ParseInt(string(fields[startTimeIndex]), 10, 64)

	return
}
