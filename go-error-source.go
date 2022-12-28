/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type GoErrorContext uint8

const none = "NONE"

var goErrorMap = map[GoErrorContext]string{
	GeNonFatal:    "GeNonFatal",
	GePreDoneExit: "GePreDoneExit",
	GeExit:        "GeExit",
}

func (ge GoErrorContext) String() (s string) {
	var ok bool
	if s, ok = goErrorMap[ge]; !ok {
		s = none
	}

	return
}
