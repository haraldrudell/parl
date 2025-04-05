/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import "github.com/haraldrudell/parl/pruntime"

type DWrapper struct {
	log func(format string, a ...any)
}

func NewDWrapper(log func(format string, a ...any), fieldp ...*DWrapper) (w *DWrapper) {

	// get w
	if len(fieldp) > 0 {
		w = fieldp[0]
	}
	if w == nil {
		w = &DWrapper{}
	}

	w.log = log
	return
}

func (w *DWrapper) D(format string, a ...any) {

	var s = pruntime.AppendLocation(
		Sprintf(format, a...),
		pruntime.NewCodeLocation(dSkipFrames),
	)
	w.log(s)
}

const (
	dSkipFrames = 1
)
