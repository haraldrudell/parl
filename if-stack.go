/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/pruntime"

type Stack interface {
	ID() ThreadID
	Status() ThreadStatus
	IsMain() (isMain bool)
	Frames() (frames []Frame)
	Creator() (creator *pruntime.CodeLocation)
	String() (s string)
}

type Frame interface {
	Loc() (location *pruntime.CodeLocation)
	Args() (args string)
	String() (s string)
}
