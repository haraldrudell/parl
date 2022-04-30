/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

var goErrorMap = map[GoErrorSource]string{
	GeNonFatal:    "GeNonFatal",
	GePreDoneExit: "GePreDoneExit",
	GeExit:        "GeExit",
}

func (ge GoErrorSource) String() (s string) {
	return goErrorMap[ge]
}
