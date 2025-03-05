/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestGoResult(t *testing.T) {
	const (
		isNewGoResult = 0
		errorMessage  = "text"
	)
	type test struct {
		label     string
		expString string
	}
	var tests = []test{{
		label:     "NewGoResult",
		expString: "goResult_len:0",
	}, {
		label:     "NewGoResult2",
		expString: "goResult_adds:0_sends:0_ch:0(1)_isError:false",
	}}
	var (
		isValid bool
		s       string
		err     error
	)

	for i, p := range tests {

		// Count() IsError() IsValid() ReceiveError()
		// Remaining() SendError() SetIsError() String()
		var goResult GoResult
		if i == isNewGoResult {
			goResult = NewGoResult()
		} else {
			goResult = NewGoResult2()
		}

		// isValid for GoResult should be true
		isValid = goResult.IsValid()
		if !isValid {
			t.Errorf("%s IsValid false", p.label)
		}

		// String should…
		s = goResult.String()
		var _ = GoResult.String
		var _ = goResultChan.String
		//var _ = goResultStruct.String
		if s != p.expString {
			t.Errorf("%s String: %q exp %q", p.label, s, p.expString)
		}

		// a failing goroutine should return error
		err = runGoroutine(errorMessage)
		if err == nil {
			t.Error("err missing")
		} else if message := err.Error(); message != errorMessage {
			t.Errorf("err message: %q exp %q", message, errorMessage)
		}
	}
}

func TestGoResultInvalid(t *testing.T) {
	const (
		expString = "GoResult_nil"
	)
	var (
		isValid bool
		s       string
	)

	// Count() IsError() IsValid() ReceiveError()
	// Remaining() SendError() SetIsError() String()
	var goResultNil GoResult

	// isValid for uninitialized GoResult should be false
	isValid = goResultNil.IsValid()
	if isValid {
		t.Error("IsValid true")
	}

	// String should…
	s = goResultNil.String()
	if s != expString {
		t.Errorf("String: %q exp %q", s, expString)
	}
}

// runGoroutine obtains error via GoResult from a failing goroutine
//   - launches and awaits a goroutine using goroutine function
//   - text: text for error message
//   - err: error returned by exiting goroutine
func runGoroutine(text string) (err error) {
	var g = NewGoResult()
	defer g.ReceiveError(&err)

	go goroutine(text, g)

	return
}

// goroutine is a goroutine function exiting with error
func goroutine(text string, g GoResult) {
	var err error
	defer g.Done(&err)
	defer RecoverErr(func() DA { return A() }, &err)

	err = perrors.New(text)
}
