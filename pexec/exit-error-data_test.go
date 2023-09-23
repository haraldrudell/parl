/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"os/exec"
	"strings"
	"testing"
)

func TestExitErrorString(t *testing.T) {
	var name = "date"
	var args = []string{"%"}
	var expStderr = "date:"

	var stdout []byte
	var err error
	var errS, stderrS string

	stdout, err = exec.Command(name, args...).Output()
	if err == nil {
		t.Fatal("exec.Command: missing error")
	}
	t.Logf("err type: %T", err)
	if e := err.(*exec.ExitError); e != nil {
		s := string(e.Stderr)
		t.Logf("err.Stderr length: %d", len(s))
	}
	if len(stdout) > 0 {
		t.Errorf("stdout output: %q", string(stdout))
	}
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
	t.Logf("stderr: %d bytes", len(stderr))
	stderrS = string(stderr)
	if !strings.HasPrefix(stderrS, expStderr) {
		t.Fatalf("stderr no prefix: %q: %q", expStderr, stderrS)
	}

	errS = NewExitErrorData(err).ExitErrorString(ExitErrorIncludeStderr)
	if !strings.Contains(errS, stderrS) {
		t.Errorf("ExitErrorString no contain %q: %q", stderrS, errS)
	}
}
