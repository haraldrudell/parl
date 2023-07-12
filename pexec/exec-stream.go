/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"context"
	"errors"
	"io"
)

var ErrArgsListEmpty = errors.New("args list empty")

// ExecStream executes a system command using the exec.Cmd type and flexible streaming.
//   - ExecStream blocks during command execution
//   - ExecStream returns any error occurring during launch or execution including
//     errors in copy threads
//   - successful exit is: statusCode == 0, isCancel == false, err == nil
//   - statusCode may be set by the process but is otherwise:
//   - — 0 successful exit
//   - — -1 process was killed by signal such as ^C or SIGTERM
//   - context cancel exit is: statusCode == -1, isCancel == true, err == nil
//   - failure exit is: statusCode != 0, isCancel == false, err != nil
//   - —
//   - args is the command followed by arguments.
//   - args[0] must specify an executable in the file system.
//     env.PATH is used to resolve the command executable
//   - if stdin stdout or stderr are nil, the are /dev/null
//     Additional threads are used to copy data when stdin stdout or stderr are non-nil
//   - os.Stdin os.Stdout os.Stderr can be provided
//   - for stdout and stderr pio has usable types:
//   - — pio.NewWriteCloserToString
//   - — pio.NewWriteCloserToChan
//   - — pio.NewWriteCloserToChanLine
//   - — pio.NewReadWriteCloserSlice
//   - any stream provided is not closed. However, upon return from ExecStream all i/o operations
//     have completed and streams may be closed as the case may be
//   - ctx is used to kill the process (by calling os.Process.Kill) if the context becomes
//     done before the command completes on its own
//   - use ExecStream with parl.EchoModerator
//   - if system commands slow down or lock-up, too many (dozens) invoking goroutines
//     may cause increased memory consumption, thrashing or exhaust of file handles, ie.
//     an uncontrollable host state
func ExecStream(stdin io.Reader, stdout io.WriteCloser, stderr io.WriteCloser,
	ctx context.Context, args ...string) (statusCode int, isCancel bool, err error) {
	return ExecStreamFull(stdin, stdout, stderr, nil, ctx, nil, nil, args...)
}
