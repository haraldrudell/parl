/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

// ThreadData contains identifiable information about a running thread.
//   - ThreadData does not have initialization
type ThreadData struct {
	// threadID is the ID of the running thread
	threadID parl.ThreadID
	// createLocation is the code line of the go-statement function-call
	// invocation launching the thread
	createLocation pruntime.CodeLocation
	// funcLocation is the code line of the function of the running thread.
	funcLocation pruntime.CodeLocation
	// optional thread-name assigned by consumer
	label string
}

var _ parl.ThreadData = &ThreadData{}

// Update populates this object from a stack trace.
func (td *ThreadData) Update(stack parl.Stack, label string) {
	td.threadID = stack.ID()
	td.createLocation = *stack.Creator()
	td.funcLocation = *stack.Frames()[len(stack.Frames())-1].Loc()
	td.label = label
}

// SetCreator gets preliminary Go identifier: the line invoking Go()
func (td *ThreadData) SetCreator(cl *pruntime.CodeLocation) {
	td.createLocation = *cl
}

// threadID is the ID of the running thread
func (td *ThreadData) ThreadID() (threadID parl.ThreadID) {
	return td.threadID
}

func (td *ThreadData) Create() (createLocation *pruntime.CodeLocation) {
	return &td.createLocation
}

func (td *ThreadData) Func() (funcLocation *pruntime.CodeLocation) {
	return &td.funcLocation
}

func (td *ThreadData) Name() (label string) {
	return td.label
}

func (td *ThreadData) Get() (threadID parl.ThreadID, createLocation pruntime.CodeLocation,
	funcLocation pruntime.CodeLocation, label string) {
	threadID = td.threadID
	createLocation = td.createLocation
	funcLocation = td.funcLocation
	label = td.label
	return
}

func (td *ThreadData) Short() (s string) {
	if td == nil {
		return "threadData:nil"
	}
	if td.label != "" {
		s = td.label
	}
	if td.threadID.IsValid() {
		if s != "" {
			s += ":" + td.threadID.String()
		} else {
			s = td.threadID.String()
		}
	}
	if s != "" {
		return
	}
	return td.String()
}

func (td *ThreadData) String() (s string) {
	var sList []string
	var s1 string
	if td.label != "" {
		s1 = td.label
	}
	if td.threadID.IsValid() {
		if s1 != "" {
			s1 += ":" + td.threadID.String()
		} else {
			s1 = td.threadID.String()
		}
	}
	if s1 != "" {
		sList = append(sList, s1)
	}
	if td.funcLocation.IsSet() {
		sList = append(sList, "func:"+td.funcLocation.Short())
	}
	if td.createLocation.IsSet() {
		sList = append(sList, "cre:"+td.createLocation.Short())
	}
	if s = strings.Join(sList, "_"); s == "" {
		s = "[empty]"
	}
	return
}
