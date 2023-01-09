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
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pio"
)

func TestExecStream(t *testing.T) {
	messageNotFound := "executable file not found"
	stdout := pio.NewWriteCloserToString()
	stderr := pio.NewWriteCloserToString()
	ctx := context.Background()
	setCommand := []string{"set"}
	sleepCommand := []string{"sleep", "1"}

	var err error
	var isCancel bool
	var statusCode int

	// empty args list
	_, _, err = ExecStream(pio.EofReader, stdout, stderr, ctx)
	if err == nil {
		t.Error("ExecStream missing err")
	} else if !errors.Is(err, ErrArgsListEmpty) {
		t.Errorf("ExecStream bad err: %q exp: %q", err.Error(), ErrArgsListEmpty)
	}

	// bash built-in: error
	_, _, err = ExecStream(pio.EofReader, stdout, stderr, ctx, setCommand...)
	if err == nil {
		t.Error("ExecStream missing err")
	} else if !strings.Contains(err.Error(), messageNotFound) {
		t.Errorf("ExecStream bad err: %q exp: %q", err.Error(), messageNotFound)
	}

	// terminate using context
	ctxCancel := parl.NewCancelContext(context.Background())
	startCallback := func(err error) {
		if err == nil {
			t.Log("startCallback invoking cancel")
			parl.InvokeCancel(ctxCancel)
		} else {
			t.Errorf("startCallback had error: %v", err)
		}
	}
	statusCode, isCancel, err = ExecStreamFull(pio.EofReader, stdout, stderr, nil, ctxCancel, startCallback, nil, sleepCommand...)
	t.Logf("Context cancel: status code: %d isCancel: %t, err: %s", statusCode, isCancel, perrors.Short(err))
	if err != nil {
		t.Errorf("ExecStream canceled context produced error: %v", err)
	} else if !isCancel {
		t.Error("ExecStream canceled context returned isCancel false")
	}
	//t.Fail()
}

// ITEST= go test ./pexec
func TestExecStreamGoodExit(t *testing.T) {
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("skiop because ITEST not set")
	}
	var args []string = []string{"sleep", "0"}

	var stdout, stderr io.WriteCloser
	var err error
	var isCancel bool
	var statusCode int
	var ctx context.Context = context.Background()
	var startCallback func(err error)

	statusCode, isCancel, err = ExecStreamFull(pio.EofReader, stdout, stderr, nil, ctx, startCallback, nil, args...)

	// Success: status code: 0 isCancel: false, err: OK
	t.Logf("Success: status code: %d isCancel: %t, err: %s", statusCode, isCancel, perrors.Short(err))
}

func TestExecStreamControlBreak(t *testing.T) {
}
