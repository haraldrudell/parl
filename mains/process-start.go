/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"os"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

const (
	kernProcPid = "kern.proc.pid"
)

// ProcessStartTime returns the time the executing process was started.
// Resolution is seconds, time zone is local
func ProcessStartTime() (createTime time.Time) {

	var unixKinfoProc *unix.KinfoProc
	var err error
	if unixKinfoProc, err = unix.SysctlKinfoProc(kernProcPid, os.Getpid()); perrors.Is(&err, "unix.SysctlKinfoProc: %T %+[1]v", err) {
		panic(err)
	}
	var unixTimeval unix.Timeval = unixKinfoProc.Proc.P_starttime
	sec, nsec := unixTimeval.Unix()
	createTime = time.Unix(sec, nsec)
	return
}
