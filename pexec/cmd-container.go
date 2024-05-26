/*
© 2024–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package pexec

import (
	"os/exec"
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// CmdContainer is a thread-safe container for the exec.Cmd value
type CmdContainer struct {
	// the command launched
	cmd atomic.Pointer[exec.Cmd]
	// any error from [exec.Cmd.Start]
	err atomic.Pointer[error]
	// triggers by [exec.Cmd.Start] callback
	awaitable parl.Awaitable
	// startResult is StartCallback
	startResult startReceiver
}

// startReceiver implements [StartCallback]
type startReceiver struct{ cmdContainer *CmdContainer }

// NewCmdContainer returns a thread-safe container for the [exec.Cmd] value
//   - startCallback: callback for [ExecStreamFull]
//   - [CmdContainer.Cmd] returns the exec.Cmd value
//   - [CmdContainer.Ch] awaits the exec.Cmd value
//   - — does not trigger if [pexec.ExecStreamFull] fails prior to [exec.Cmd.Start]
//   - [CmdContainer.Err] returns any error from [exec.Cmd.Start]
func NewCmdContainer() (cmd *CmdContainer, startCallback StartCallback) {
	cmd = &CmdContainer{}
	cmd.startResult.cmdContainer = cmd
	startCallback = &cmd.startResult
	return
}

// StartResult stores the exec.Cmd value as soon as the process has started
func (s *startReceiver) StartResult(execCmd *exec.Cmd, err error) {
	c := s.cmdContainer
	if c.awaitable.IsClosed() {
		return
	} else if err != nil {
		c.err.Store(&err)
	} else {
		c.cmd.Store(execCmd)
	}
	c.awaitable.Close()
}

// Ch awaits process start when the Cmd value is available
//   - does not trigger if ExecStreamFull fails prior to [exec.Cmd.Start]
func (c *CmdContainer) Ch() (ch parl.AwaitableCh) { return c.awaitable.Ch() }

// Cmd returns the exec.Cmd value, available after process start
func (c *CmdContainer) Cmd() (execCmd *exec.Cmd) { return c.cmd.Load() }

// Cmd returns any error from [exec.Cmd.Start]
func (c *CmdContainer) Err() (err error) {
	if ep := c.err.Load(); ep != nil {
		err = *ep
	}
	return
}
