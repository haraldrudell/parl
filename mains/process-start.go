/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"fmt"
	"os"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/shirou/gopsutil/process"
)

// ProcessStartTime returns the time the executing process was started.
// Resolution is seconds, time zone is local
func ProcessStartTime() (createTime time.Time) {

	// get process object for this process
	var procData *process.Process
	pid := os.Getpid()
	var err error
	if procData, err = process.NewProcess(int32(pid)); err != nil {
		panic(perrors.NewPF(fmt.Sprintf("process.NewProcess(%d) error: %T %+[2]v", pid, err)))
	}

	// get process create time in seconds
	var epochMs int64
	if epochMs, err = procData.CreateTime(); err != nil {
		panic(perrors.ErrorfPF("Process.CreateTime error: %T %+[1]v", err))
	}
	createTime = time.UnixMilli(epochMs).Local()

	return
}
