/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
)

// WarningType is an error with lesser impact
type WarningType struct {
	ErrorChain
}

var _ error = &WarningType{}

// Warning indicates a problem of less severity than error
func Warning(err error) error {
	return Stack(&WarningType{ErrorChain{err}})
}

// IsWarning determines if an error is a warning
func IsWarning(err error) bool {
	var warning WarningType
	return errors.Is(err, warning)
}

func (w WarningType) Error() (s string) {
	if err := w.Unwrap(); err != nil {
		return err.Error()
	}
	return "warning"
}
