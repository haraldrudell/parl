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

// ExecBlocking provides a simple way to quietly execute a system command
//   - blocking, control over standard i/o, ability to cancel
//   - stdin: a byte sequence used as input to the system command
//   - stdin ExecBlockingClosedStdin nil: the command receives a closed standard input
//   - stdin ExecBlockingEmptyStdin: the system command receives an open but empty standard input
//   - wantStdout WantStdout: the command’s standard poutput is returned in stdout
//   - wantStdout NoExecBlockingStdout: the command’s standard output is discarded
//   - wantStderr WantStderr: the command’s standard error output is returned in stderr
//   - wantStderr StderrIsError: if the command outputs to standard error, the output is returned as an error
//   - wantStderr NoExecBlockingStderr: the command’s standard error output is discarded.
//     standard error output may still be provided in err
//   - ctx: termianes the commad usinng SIGKILL on context cancel. Cannot be nil
//   - args: args[0] is non-empty command. args[1…] holds command-line arguments
func ExecBlocking(stdin []byte, wantStdout Stdouts, wantStderr Stderrs, ctx context.Context, args ...string) (stdout, stderr *bytes.Buffer, err error) {
	if len(args) == 0 || args[0] == "" {
		err = perrors.ErrorfPF("%w", ErrArgsListEmpty)
		return
	}

	// execCmd contains launch values for a sub-process executing the system command
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

	// run command: blocks here
	err = execCmd.Run()

	// error executing command or starting the process
	if err != nil {
		err = perrors.ErrorfPF("system command failed: %s error: “%w” stderr: %q",
			strings.Join(args, "\x20"), err, stderr0.String(),
		)
	} else if wantStderr == StderrIsError && stderr0.Len() > 0 {
		err = perrors.ErrorfPF("system command wrote to standard error: %s stderr: %q",
			strings.Join(args, "\x20"), stderr0.String(),
		)
	}

	return
}

const (
	// [ExecBlocking] system command’s standard output is discarded
	NoExecBlockingStdout Stdouts = false
	// [ExecBlocking] system command’s standard output is be captured
	WantStdout Stdouts = true
)

// Stdouts are the possible values for [ExecBlocking] stdout
//   - NoExecBlockingStdout WantStdout
type Stdouts bool

const (
	// [ExecBlocking] system command’s standard error is discarded
	NoExecBlockingStderr Stderrs = iota + 1
	// [ExecBlocking] system command’s standard error is captured
	WantStderr
	// [ExecBlocking] any system command output to standard error is returned as an error
	StderrIsError
)

// Stderrs are the possible values for [ExecBlocking] stderr
//   - NoExecBlockingStderr WantStderr StderrIsError
type Stderrs uint8

// [ExecBlocking] the system command receives a closed standard input
var NoExecBlockingStdin []byte

// [ExecBlocking] the system command receives an open but empty standard input
var ExecBlockingEmptyStdin = []byte{}

// [ExecBlocking] the system command receives a closed standard input
var ExecBlockingClosedStdin = []byte{}
