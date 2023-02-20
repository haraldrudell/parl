/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"strings"

	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	puFrames = 1
)

// error116.Warning indicates that err is a problem of less severity than error.
// It is uesed for errors that are not to terminate the thread.
// A Warning can be detected using error116.IsWarning().
func Warning(err error) error {
	return Stack(errorglue.NewWarning(err))
}

// error116.AddKeyValue attaches a string value to err.
// The values can be trrioeved using error116.ErrorData().
// if key is non-empty valiue is returned in a map where last key wins.
// if key is empty, valuse is returned in s string slice.
// err can be nil.
func AddKeyValue(err error, key, value string) (e error) {
	return errorglue.NewErrorData(err, key, value)
}

// error116.AppendError associates an additional error with err.
// err and err2 can be nil.
// Associated error instances can be retrieved using error116.AllErrors, error116.ErrorList or
// by printing using rich error printing of the error116 package.
// TODO 220319 fill in printing
func AppendError(err error, err2 error) (e error) {
	if err2 == nil {
		return err // noop return
	}
	if err == nil {
		return err2 // single error return
	}
	return errorglue.NewRelatedError(err, err2)
}

// AppendErrorDefer aggregates error sources into errp.
//   - AppendErrorDefer is deferrable
//   - errp cannot be nil
//   - errp2 is a pointer to a second error variable used as source.
//     If errp2 is nil or *errp2 is nil, no action is taken
//   - fn is a function returning a possible error.
//     If fn is nil or fn returns nil, no action is taken
//   - AppendErrorDefer uses AppendError to aggregate error values into *errp
func AppendErrorDefer(errp, errp2 *error, fn func() (err error)) {
	if errp == nil {
		panic(NewPF("errp cannot be nil"))
	}
	if errp2 != nil {
		if err := *errp2; err != nil {
			*errp = AppendError(*errp, err)
		}
	}
	if fn != nil {
		if err := fn(); err != nil {
			*errp = AppendError(*errp, err)
		}
	}
}

func TagErr(e error, tags ...string) (err error) {

	// ensure error has stack
	if !HasStack(e) {
		e = Stackn(e, puFrames)
	}

	// values to print
	s := pruntime.NewCodeLocation(puFrames).PackFunc()
	if tagString := strings.Join(tags, "\x20"); tagString != "" {
		s += "\x20" + tagString
	}

	return Errorf("%s: %w", s, e)
}

func InvokeIfError(errp *error, errFn func(err error)) {
	var err error
	if errp != nil {
		err = *errp
	} else {
		err = New("perrors.InvokeIfError errp nil")
	}
	if err != nil {
		errFn(err)
	}
}
