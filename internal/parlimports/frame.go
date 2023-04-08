/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlimports

import (
	"path/filepath"
	"strconv"

	"github.com/haraldrudell/parl/pruntime"
)

type Frame struct {
	pruntime.CodeLocation
	// args like "(1, 2, 3)"
	args string
}

func (f *Frame) Loc() (location *pruntime.CodeLocation) {
	return &f.CodeLocation
}

func (f *Frame) Args() (args string) {
	return f.args
}

func (f *Frame) String() (s string) {
	return f.CodeLocation.FuncName + f.args + "\n" +
		filepath.Base(f.CodeLocation.File) + ":" + strconv.Itoa(f.CodeLocation.Line)
}
