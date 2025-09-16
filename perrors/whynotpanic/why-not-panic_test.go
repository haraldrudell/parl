/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package whynotpanic

import (
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

func TestWhyNotPanic(t *testing.T) {
	var errorNilExp = "error-nil: true"
	var errorNoStackExp = "stacks: 0"
	var errorStackNoPanicExp = "stacks: 1"

	var s string
	var errorNil error
	var errorNoStack = errors.New("errorNoStack")
	var errorStackNoPanic = errorglue.NewErrorStack(errors.New("errorStackNoPanic"), pruntime.NewStack(0))
	var _, isErrorCallStacker = errorStackNoPanic.(errorglue.ErrorCallStacker)
	var errorIsPanic = recoverPanic1()
	var errorPanicWithStack = recoverPanic2()

	t.Logf("errorStackNoPanic chain: %d[%s] ErrorCallStacker: %t",
		len(errorChainSlice(errorStackNoPanic)),
		errorglue.DumpChain(errorStackNoPanic),
		isErrorCallStacker,
	)

	// error nil should return error-nil: true
	s = WhyNotPanic(errorNil)
	if !strings.Contains(s, errorNilExp) {
		t.Errorf("nil error not detected: %s", s)
	}

	// error without stack should return stacks: 0
	s = WhyNotPanic(errorNoStack)
	if !strings.Contains(s, errorNoStackExp) {
		t.Errorf("errorNoStackExp has stacks: %s", s)
	}

	// error with stack should return stacks: 1
	s = WhyNotPanic(errorStackNoPanic)
	if !strings.Contains(s, errorStackNoPanicExp) {
		t.Errorf("errorNoStackExp has stacks: %s", s)
	}

	// a recovered panic should return empty string
	s = WhyNotPanic(errorIsPanic)
	if s != "" {
		t.Errorf("errorIsPanic no panic detected: %s", s)
	}

	// a recovered panic with an error already having a stack
	// should return empty string
	s = WhyNotPanic(errorPanicWithStack)
	if s != "" {
		t.Errorf("errorPanicWithStack no panic detected: %s", s)
	}
}

func recoverPanic1() (err error) {
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err)

	panic(1)
}

func recoverPanic2() (err error) {
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err)

	var errorWithStack = errorglue.NewErrorStack(errors.New("errorStackNoPanic"), pruntime.NewStack(0))
	panic(errorWithStack)
}

// errorChainSlice returns a slice of errors from following
// the main error chain
//   - err: an error to traverse
//   - errs: all errors in the main error chain beginning with err itself
//   - — nil if err was nil
//   - — otherwise of length 1 or more
func errorChainSlice(err error) (errs []error) {
	for err != nil {
		errs = append(errs, err)
		err, _, _ = errorglue.Unwrap(err)
	}
	return
}
