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
	//	- a small integer with 1 for main thread, displayed by debug.Stack
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
func (t *ThreadData) Update(
	threadID parl.ThreadID,
	createInvocation, goFunction *pruntime.CodeLocation,
	label string) {
	if !t.threadID.IsValid() && threadID.IsValid() {
		t.threadID = threadID
	}
	if createInvocation != nil && !t.createLocation.IsSet() && createInvocation.IsSet() {
		t.createLocation = *createInvocation
	}
	if goFunction != nil && !t.funcLocation.IsSet() && goFunction.IsSet() {
		t.funcLocation = *goFunction
	}
	if t.label == "" && label != "" {
		t.label = label
	}
}

// SetCreator gets preliminary Go identifier: the line invoking Go()
func (t *ThreadData) SetCreator(cl *pruntime.CodeLocation) {
	t.createLocation = *cl
}

// threadID is the ID of the running thread
func (t *ThreadData) ThreadID() (threadID parl.ThreadID) {
	return t.threadID
}

func (t *ThreadData) Create() (createLocation *pruntime.CodeLocation) {
	return &t.createLocation
}

func (t *ThreadData) Func() (funcLocation *pruntime.CodeLocation) {
	return &t.funcLocation
}

func (t *ThreadData) Name() (label string) {
	return t.label
}

func (t *ThreadData) Get() (threadID parl.ThreadID, createLocation pruntime.CodeLocation,
	funcLocation pruntime.CodeLocation, label string) {
	threadID = t.threadID
	createLocation = t.createLocation
	funcLocation = t.funcLocation
	label = t.label
	return
}

// "myThreadName:4"
func (t *ThreadData) Short() (s string) {

	// handle nil case
	if t == nil {
		return ThreadDataNil // "threadData:nil"
	}

	// "[label]:[threadID]"
	if t.label != "" {
		s = t.label
	}
	if t.threadID.IsValid() {
		if s != "" {
			s += ":" + t.threadID.String()
		} else {
			s = t.threadID.String()
		}
	}
	if s != "" {
		return
	}

	// zero-value case
	return t.String()
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
