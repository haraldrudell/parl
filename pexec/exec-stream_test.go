/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pio"
	"golang.org/x/sys/unix"
)

func TestExecStream(t *testing.T) {
	var (
		messageNotFound = "executable file not found"
		setCommand      = []string{"set"}
		sleepCommand    = []string{"sleep", "1"}
	)

	var (
		err                 error
		isCancel            bool
		statusCode          int
		cancelingContext    context.Context
		startCallback       StartCallback
		stdoutWriteCloser   = pio.NewCloserBuffer()
		stderrWriteCloser   = pio.NewCloserBuffer()
		nonCancelingContext = context.Background()
	)

	// empty args list should return error
	statusCode, isCancel, err = ExecStream(pio.EofReader, stdoutWriteCloser, stderrWriteCloser, nonCancelingContext)
	_ = statusCode
	_ = isCancel
	if err == nil {
		t.Error("ExecStream missing err")
	} else if !errors.Is(err, ErrArgsListEmpty) {
		t.Errorf("ExecStream bad err: %q exp: %q", perrors.Short(err), ErrArgsListEmpty)
	}

	// executing a bash built-in as command should return error
	statusCode, isCancel, err = ExecStream(pio.EofReader, stdoutWriteCloser, stderrWriteCloser, nonCancelingContext, setCommand...)
	_ = statusCode
	_ = isCancel
	if err == nil {
		t.Error("ExecStream missing err")
	} else if !strings.Contains(err.Error(), messageNotFound) {
		t.Errorf("ExecStream bad err: %q exp: %q", perrors.Short(err), messageNotFound)
	}

	// canceling context should terminate the sub-process
	cancelingContext = parl.NewCancelContext(context.Background())
	startCallback = newCancelingStartCallback(t, cancelingContext)
	statusCode, isCancel, err = ExecStreamFull(pio.EofReader, stdoutWriteCloser, stderrWriteCloser, nil, cancelingContext, startCallback, nil, sleepCommand...)
	t.Logf("ExecStreamFull returned values on context cancel: status code: %d isCancel: %t, err: %s", statusCode, isCancel, perrors.Short(err))
	if err != nil {
		t.Errorf("ExecStream canceled context produced error: %s", perrors.Long(err))
	} else if !isCancel {
		t.Error("ExecStream canceled context returned isCancel false")
	}
}

// ITEST= go test ./pexec
func TestExecStreamGoodExit(t *testing.T) {

	// this integration test only runs if environment variable ITEST= is set
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("skip because ITEST not set")
	}

	var (
		// true is a successful command exiting immediately
		args []string = []string{"true"}
	)

	var (
		stdout, stderr io.WriteCloser
		err            error
		isCancel       bool
		statusCode     int
		ctx            context.Context = context.Background()
		startCallback  StartCallback
	)

	statusCode, isCancel, err = ExecStreamFull(pio.EofReader, stdout, stderr, nil, ctx, startCallback, nil, args...)

	// Success: status code: 0 isCancel: false, err: OK
	t.Logf("Success: status code: %d isCancel: %t, err: %s", statusCode, isCancel, perrors.Short(err))
}

// ITEST= go test -v -run ^TestExecStreamSIGTERM$ github.com/haraldrudell/parl/pexec
//
// exec-stream_test.go:112: Success: status code: -1 isCancel: false, err: pexec.ExecStreamFull execCmd.Wait signal: terminated at pexec.ExecStreamFull()-exec-stream-full.go:168
func TestExecStreamSIGTERM(t *testing.T) {

	// this integration test only runs if environment variable ITEST= is set
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("skip because ITEST not set")
	}

	var (
		args []string = []string{"sleep", "10"}
	)

	var (
		stdout, stderr io.WriteCloser
		err            error
		isCancel       bool
		statusCode     int
		ctx            context.Context = context.Background()
	)

	statusCode, isCancel, err = ExecStreamFull(pio.EofReader, stdout, stderr, nil, ctx, &sigTERMStartCallback{}, nil, args...)

	// Success: status code: 0 isCancel: false, err: OK
	t.Logf("Success: status code: %d isCancel: %t, err: %s", statusCode, isCancel, perrors.Short(err))
}

// sigTERMStartCallback is a [StartCallback] immediately sends signal SIGTERM
type sigTERMStartCallback struct{}

// StartResult immediately sends signal SIGTERM
func (s *sigTERMStartCallback) StartResult(execCmd *exec.Cmd, err error) {
	execCmd.Process.Signal(unix.SIGTERM)
}

// cancelingStartCallback is a [StartCallback] immediately canceling cancelingContext
type cancelingStartCallback struct {
	t                *testing.T
	cancelingContext context.Context
}

// newCancelingStartCallback returns a [StartCallback] immediately canceling cancelingContext
func newCancelingStartCallback(t *testing.T, cancelingContext context.Context) (s *cancelingStartCallback) {
	return &cancelingStartCallback{
		t:                t,
		cancelingContext: cancelingContext,
	}
}

// StartResult immediately cancels cancelingContext
func (s *cancelingStartCallback) StartResult(execCmd *exec.Cmd, err error) {
	t := s.t
	if err == nil {
		t.Log("startCallback invoking cancel")
		parl.InvokeCancel(s.cancelingContext)
	} else {
		t.Errorf("startCallback had error: %s", perrors.Short(err))
	}
}
