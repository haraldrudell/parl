/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/pbytes"
)

func TestExitErrorString(t *testing.T) {
	var name = "date"
	var args = []string{"%"}
	var expStderr = "date:"

	var stdout, exitErrorStderr []byte
	var err error
	var errS, stderrS string

	// command: “date %” produces an error
	stdout, err = exec.Command(name, args...).Output()
	if err == nil {
		t.Fatal("exec.Command: missing error")
	}

	// verify error type to be exec.ExitError
	//	- those errors have status code and stderr
	// err type: *exec.ExitError
	t.Logf("err type: %T", err)

	// command should echo to stderr
	if e := err.(*exec.ExitError); e != nil {
		exitErrorStderr = e.Stderr
		t.Logf("err.Stderr length: %d", len(string(exitErrorStderr)))
		if len(exitErrorStderr) == 0 {
			t.Error("stderrr empty")
		}
	}

	// command should not echo to stdout
	if len(stdout) > 0 {
		t.Errorf("stdout output: %q", string(stdout))
	}

	// status code should be non-zero and no signal
	hasStatusCode, statusCode, signal, stderr := ExitError(err)
	if !hasStatusCode {
		t.Fatal("err not ExitError")
	}
	if signal != 0 {
		t.Errorf("signal: %s", signal)
	}
	if statusCode == 0 {
		t.Error("status code: 0")
	}

	// returned stderr should contain expStderr
	t.Logf("stderr: %d bytes", len(stderr))
	stderrS = string(stderr)
	if !strings.HasPrefix(stderrS, expStderr) {
		t.Fatalf("stderr no prefix: %q: %q", expStderr, stderrS)
	}

	errS = NewExitErrorData(err).ExitErrorString(ExitErrorIncludeStderr)
	expErrS := string(pbytes.TrimNewline(exitErrorStderr))
	if !strings.Contains(errS, expErrS) {
		t.Errorf("ExitErrorString no contain %q: %q", stderrS, expErrS)
	}
}
