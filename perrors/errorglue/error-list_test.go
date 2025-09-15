/*
© 2012–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestErrorListAppend(t *testing.T) {
	// t.Error("Logging on")
	const (
		one, two, three, four, five, six = "1", "2", "3", "4", "5", "6"
		aLetter, bLetter, rLetter        = "A", "B", "R"
		// “64”
		//	- graph: 6 → 5 → append(a → 2 → 1, b → 4 → 3)
		//	- only one additional error-chain: 4
		exp = "64"
	)
	var (
		err1, errA, err2, err3, errB, err4, errR, err5, err6 error
		errMap                                               map[error]string
		errs                                                 []error
		actual, eS                                           string
	)

	// build:
	//	- 6 → 5 → append(a → 2 → 1, b → 4 → 3)
	err1 = errorf(one)
	errA, _, _ = Unwrap(err1)
	t.Logf("%s", errorGraph(err1))
	err2 = errorf("%s %w", two, err1)
	t.Logf("%s", errorGraph(err2))
	err3 = errorf(three)
	errB, _, _ = Unwrap(err3)
	err4 = errorf("%s %w", four, err3)
	errR = appendError(err2, err4)
	t.Logf("%s", errorGraph(errR))
	err5 = errorf("%s %w", five, errR)
	t.Logf("%s", errorGraph(err5))
	err6 = errorf("%s %w", six, err5)
	errMap = map[error]string{
		err1: one,
		errA: aLetter,
		err2: two,
		err3: three,
		errB: bLetter,
		err4: four,
		errR: rLetter,
		err5: five,
		err6: six,
	}

	// *fmt.wrapError"6 5 2 1" → *fmt.wrapError"5 2 1" →
	// *errorglue.relatedError"2 1"
	// → (a → *fmt.wrapError"2 1" → *errorglue.errorStack"1" → *errors.errorString"1",
	// b → *fmt.wrapError"4 3" → *errorglue.errorStack"3" → *errors.errorString"3")
	t.Logf("%s", errorGraph(err6))

	t.Log("error graph: 6 → 5 → R(a → 2 → 1A, b → 4 → 3B)")

	errs = ErrorList(err6)

	// errs[0]: "6 5 2 1"
	t.Logf("len(errs): %d errs[0].Errror(): %q", len(errs), errs[0].Error())

	if len(errs) != len(exp) {
		t.Errorf("FAIL bad errs length: %d exp %d", len(errs), len(exp))
	}

	actual = ""
	for _, e := range errs {
		eS = errMap[e]
		if eS != "" {
			actual += eS
		} else {
			t.Errorf("FAIL unknown error: %T ‘%[1]v’", e)
		}
	}

	// actual: [6 5 2 1 4 3]
	t.Logf("actual error list: %s", actual)

	if actual != exp {
		t.Errorf("FAIL %q exp %q", actual, exp)
	}
}

func TestErrorListJoin(t *testing.T) {
	// t.Error("Logging on")
	const (
		one, two, three, four, five, six = "1", "2", "3", "4", "5", "6"
		aLetter, bLetter, rLetter        = "A", "B", "R"
		exp                              = "64"
	)
	var (
		err1, errA, err2, err3, errB, err4, errR, err5, err6 error
		errMap                                               map[error]string
		errs                                                 []error
		actual, eS                                           string
	)

	// build:
	//	- 6 → 5 → append(a → 2 → 1, b → 4 → 3)
	err1 = errorf(one)
	errA, _, _ = Unwrap(err1)
	t.Logf("%s", errorGraph(err1))
	err2 = errorf("%s %w", two, err1)
	t.Logf("%s", errorGraph(err2))
	err3 = errorf(three)
	errB, _, _ = Unwrap(err3)
	err4 = errorf("%s %w", four, err3)
	errR = errors.Join(err2, err4)
	t.Logf("%s", errorGraph(errR))
	err5 = errorf("%s %w", five, errR)
	t.Logf("%s", errorGraph(err5))
	err6 = errorf("%s %w", six, err5)
	errMap = map[error]string{
		err1: one,
		errA: aLetter,
		err2: two,
		err3: three,
		errB: bLetter,
		err4: four,
		errR: rLetter,
		err5: five,
		err6: six,
	}

	// *fmt.wrapError"6 5 2 1" → *fmt.wrapError"5 2 1" →
	// *errorglue.relatedError"2 1"
	// → (a → *fmt.wrapError"2 1" → *errorglue.errorStack"1" → *errors.errorString"1",
	// b → *fmt.wrapError"4 3" → *errorglue.errorStack"3" → *errors.errorString"3")
	t.Logf("%s", errorGraph(err6))

	t.Log("error graph: 6 → 5 → R(a → 2 → 1A, b → 4 → 3B)")

	errs = ErrorList(err6)

	// errs[0]: "6 5 2 1"
	t.Logf("len(errs): %d errs[0].Errror(): %q", len(errs), errs[0].Error())

	if len(errs) != len(exp) {
		t.Errorf("FAIL bad errs length: %d exp %d", len(errs), len(exp))
	}

	actual = ""
	for _, e := range errs {
		eS = errMap[e]
		if eS != "" {
			actual += eS
		} else {
			t.Errorf("FAIL unknown error: %T ‘%[1]v’", e)
		}
	}

	// actual: [6 5 2 1 4 3]
	t.Logf("actual error list: %s", actual)

	if actual != exp {
		t.Errorf("FAIL %q exp %q", actual, exp)
	}
}

func errorf(format string, a ...interface{}) (err error) {
	err = fmt.Errorf(format, a...)
	if hasStack(err) {
		return
	}
	err = stackn(err, 1)
	return
}

func hasStack(err error) (hasStack bool) {
	if err == nil {
		return
	}
	var e ErrorCallStacker
	// if an error of type ErrorCallStacker is found in err’s error chain,
	// hasStack is true
	hasStack = errors.As(err, &e)
	return
}

// Stackn always attaches a new stack trace to non-nil err
//   - framesToSkip: 0 is caller, larger skips stack frames
func stackn(err error, framesToSkip int) (err2 error) {
	if err == nil {
		return
	} else if framesToSkip < 0 {
		framesToSkip = 0
	}
	err2 = NewErrorStack(
		err,
		pruntime.NewStack(1+framesToSkip),
	)
	return
}

func appendError(err error, err2 error) (e error) {
	if err2 == nil {
		e = err // err2 is nil, return is err, possibly nil
	} else if err == nil {
		e = err2 // err is nil, return is non-nil err2
	} else {
		e = NewRelatedError(err, err2) // both non-nil
	}
	return
}

// errorGraph displays allchains of err
//   - cyclic graphs hang
func errorGraph(err error) (s string) {

	if err == nil {
		s = "err:nil"
		return
	}

	var sL1 []string
	for err != nil {

		// print err
		sL1 = append(
			sL1,
			fmt.Sprintf("%T%q", err, err.Error()),
		)

		// handle related
		if relatedError, isRelated := err.(RelatedError); isRelated {
			var associatedErr = relatedError.AssociatedError()
			var nextErr error
			if wrappedErr, isWrap := err.(Unwrapper); isWrap {
				nextErr = wrappedErr.Unwrap()
			}

			sL1 = append(
				sL1,
				fmt.Sprintf("(a → %s, b → %s)",
					errorGraph(nextErr),
					errorGraph(associatedErr),
				),
			)
			s = strings.Join(sL1, " → ")
			return
		}

		// handle Unwrap
		switch wrapper := err.(type) {
		case Unwrapper:
			err = wrapper.Unwrap()
		case JoinUnwrapper:
			var errList = wrapper.Unwrap()
			var sL = make([]string, len(errList))
			for i, e := range errList {
				sL[i] = fmt.Sprintf("%d → %s", i, errorGraph(e))
			}
			sL1 = append(
				sL1,
				fmt.Sprintf("(%s)", strings.Join(sL, ", ")),
			)
			s = strings.Join(sL1, " → ")
			return
		default:
			err = nil
		}
	}
	s = strings.Join(sL1, " → ")

	return
}
