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

const (
	ThreadDataEmpty = "[empty]"
	ThreadDataNil   = "threadData:nil"
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
func (td *ThreadData) Update(
	threadID parl.ThreadID,
	createInvocation, goFunction *pruntime.CodeLocation,
	label string) {
	if !td.threadID.IsValid() && threadID.IsValid() {
		td.threadID = threadID
	}
	if createInvocation != nil && !td.createLocation.IsSet() && createInvocation.IsSet() {
		td.createLocation = *createInvocation
	}
	if goFunction != nil && !td.funcLocation.IsSet() && goFunction.IsSet() {
		td.funcLocation = *goFunction
	}
	if td.label == "" && label != "" {
		td.label = label
	}
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

// "myThreadName:4"
func (td *ThreadData) Short() (s string) {

	// handle nil case
	if td == nil {
		return ThreadDataNil // "threadData:nil"
	}

	// "[label]:[threadID]"
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

	// zero-value case
	return td.String()
}

func (t *ThreadData) LabeledString() (s string) {
	var sL []string
	if t.label != "" {
		sL = append(sL, "label: "+t.label)
	}
	if t.threadID.IsValid() {
		sL = append(sL, "threadID: "+t.threadID.String())
	}
	if t.funcLocation.IsSet() {
		sL = append(sL, "go-function: "+t.funcLocation.Short())
	}
	if t.createLocation.IsSet() {
		sL = append(sL, "go-statement: "+t.createLocation.Short())
	}
	if len(sL) > 0 {
		s = strings.Join(sL, "\x20")
	} else {
		s = "[no data]"
	}
	return
}

// "myThreadName:4_func:testing.tRunner()-testing.go:1446_cre:testing.(*T).Run()-testing.go:1493"
func (t *ThreadData) String() (s string) {
	var sList []string
	var s1 string
	if t.label != "" {
		s1 = t.label
	}
	if t.threadID.IsValid() {
		if s1 != "" {
			s1 += ":" + t.threadID.String()
		} else {
			s1 = t.threadID.String()
		}
	}
	if s1 != "" {
		sList = append(sList, s1)
	}
	if t.funcLocation.IsSet() {
		sList = append(sList, "func:"+t.funcLocation.Short())
	}
	if t.createLocation.IsSet() {
		sList = append(sList, "cre:"+t.createLocation.Short())
	}
	if s = strings.Join(sList, "_"); s == "" {
		s = ThreadDataEmpty
	}
	return
}
