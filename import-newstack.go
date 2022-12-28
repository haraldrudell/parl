/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

var newStack func(skipFrames int) (stack Stack)

func ImportNewStack(pdebugNewStack func(skipFrames int) (stack Stack)) {
	if newStack == nil {
		newStack = pdebugNewStack
	}
}
