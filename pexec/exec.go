/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

const (
	NoExecBlockingStdout Stdouts = false
	WantStdout           Stdouts = true
)

type Stdouts bool

const (
	NoExecBlockingStderr Stderrs = iota + 1
	WantStderr
	StderrIsError
)

type Stderrs uint8

var NoExecBlockingStdin []byte

// ExecBlocking executes a system command with independent control over stdin stdout stderr
//   - if stdin is nil, no stdin is provided to command
//   - if wantStdout is true, stdout contains any command output, if false stdout nil
//   - if wantStderr is true, stderr contains any command error output, if false stderr nil
func ExecBlocking(stdin []byte, wantStdout Stdouts, wantStderr Stderrs, ctx context.Context, args ...string) (stdout, stderr *bytes.Buffer, err error) {
	if len(args) == 0 {
		err = perrors.ErrorfPF("%w", ErrArgsListEmpty)
		return
	}

	var execCmd = exec.CommandContext(ctx, args[0], args[1:]...)
	if stdin != nil {
		execCmd.Stdin = bytes.NewReader(stdin)
	}
	if wantStdout == WantStdout {
		stdout = new(bytes.Buffer)
		execCmd.Stdout = stdout
	}
	var stderr0 bytes.Buffer
	execCmd.Stderr = &stderr0
	if wantStderr == WantStderr {
		stderr = &stderr0
	}

	// run command
	if err = execCmd.Run(); err != nil {
		err = perrors.ErrorfPF("system command failed: %s error: '%w' stderr: %q",
			strings.Join(args, "\x20"), err, stderr0.String(),
		)
	} else if wantStderr == StderrIsError && stderr0.Len() > 0 {
		err = perrors.ErrorfPF("system command wrote to standard error: %s stderr: %q",
			strings.Join(args, "\x20"), stderr0.String(),
		)
	}

	return
}
