/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"fmt"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

func GoErrorDump(goError parl.GoError) (s string) {

	// check for nil
	s = fmt.Sprintf("parl.GoError: type: %T", goError)
	if goError == nil {
		return // goError nil returns "parl.GoError: type: <nil>"
	}

	// ensure g0.GoError
	var ge *GoError
	var ok bool
	if ge, ok = goError.(*GoError); !ok {
		s += fmt.Sprintf(" type is not %T", ge)
		return
	}

	var sGo string
	if ge.g0 == nil {
		sGo = "nil"
	} else {
		sGo = ge.g0.ThreadInfo().String()
	}

	// GoError.err t errContext
	s += fmt.Sprintf(" err: %s t: %s errContext: %s Go: %s",
		perrors.Short(ge.err),
		ge.t.Format(parl.Rfc3339ns),
		ge.errContext,
		sGo,
	)

	return
}
