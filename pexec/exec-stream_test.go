/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pio"
)

func TestExecStream(t *testing.T) {
	messageNotFound := "executable file not found"
	stdout := pio.NewWriteCloserToString()
	stderr := pio.NewWriteCloserToString()
	env := []string{}
	ctx := context.Background()
	setCommand := []string{"set"}
	sleepCommand := []string{"sleep", "1"}

	var err error
	var isCancel bool

	// empty args list
	_, err = ExecStream(pio.EofReader, stdout, stderr, env, ctx, nil)
	if err == nil {
		t.Error("ExecStream missing err")
	} else if !errors.Is(err, ErrArgsListEmpty) {
		t.Errorf("ExecStream bad err: %q exp: %q", err.Error(), ErrArgsListEmpty)
	}

	// bash built-in: error
	_, err = ExecStream(pio.EofReader, stdout, stderr, nil, ctx, nil, setCommand...)
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
	isCancel, err = ExecStream(pio.EofReader, stdout, stderr, nil, ctxCancel, startCallback, sleepCommand...)
	if err != nil {
		t.Errorf("ExecStream canceled context produced error: %v", err)
	} else if !isCancel {
		t.Error("ExecStream canceled context returned isCancel false")
	}
}
