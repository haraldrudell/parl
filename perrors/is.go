/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "fmt"

const (
	isStackFrames = 1
)

func Is2(errp *error, e error, format string, a ...interface{}) (isBad bool) {
	if e == nil {
		return // no error exit
	} else if !HasStack(e) {
		e = Stackn(e, isStackFrames)
	}
	Is(&e, format, a...)
	if errp != nil {
		*errp = AppendError(*errp, e)
	}
	return true
}

func Is2PF(errp *error, e error, format string, a ...interface{}) (isBad bool) {
	if e == nil {
		return // no error exit
	} else if !HasStack(e) {
		e = Stackn(e, isStackFrames)
	}
	PF := PackFuncN(isStackFrames)
	if format == "" {
		e = fmt.Errorf("%s %w", PF, e)
	} else {
		e = fmt.Errorf("%s "+format, append([]interface{}{PF}, a...)...)
	}
	if errp != nil {
		*errp = AppendError(*errp, e)
	}
	return true
}

func Is(errp *error, format string, a ...interface{}) (isBad bool) {
	if errp == nil {
		panic(NewPF("errp nil"))
	}
	err := *errp
	if err == nil {
		return // no error exit
	}
	if format != "" {
		err = fmt.Errorf(format, a...)
	}
	if !HasStack(err) {
		err = Stackn(err, isStackFrames)
	}
	*errp = err
	return true
}

func IsPF(errp *error, format string, a ...interface{}) (isBad bool) {
	if errp == nil {
		panic(NewPF("errp nil"))
	}
	err := *errp
	if err == nil {
		return false // no error exit
	}
	PF := PackFuncN(isStackFrames)
	if format == "" {
		err = fmt.Errorf("%s %w", PF, err)
	} else {
		err = fmt.Errorf("%s "+format, append([]interface{}{PF}, a...)...)
	}
	if !HasStack(err) {
		err = Stackn(err, isStackFrames)
	}
	*errp = err
	return true
}
