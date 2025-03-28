/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/haraldrudell/parl/pruntime"
)

// Thread interface

// Go provides the four needs of a running goroutione thread.
// The Go is provided as a function argument in the go statement function call
// that launches the thread.
//   - the four needs:
//   - — to be waited upon via [Go.Done]
//   - — to submit non-fatal errors via [Go.AddError]
//   - — to detect and initiate cancel via [Go.Context] [Go.Cancel]
//   - [Go.Cancel] cancels:
//   - — this Go’s parent thread-group’s context and
//   - — this Go’s parent thread-group’s subordinate thread-groups’ contexts
//   - The Go Context is canceled when
//   - — the parent GoGroup thread-group’s context is Canceled or
//   - —a thread in the parent GoGroup thread-group initiates Cancel
//   - Cancel by threads in subordinate thread-groups do not Cancel this Go thread
//
// Usage:
//
//	var threadGroup = g0.NewGoGroup(context.Background())
//	go someFunc(threadGroup.Go())
//	…
//	func someFunc(g parl.Go) {
//	  var err error
//	  defer g.Register().Done(&err)
//	  defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)
//	  …
type Go interface {
	// Register performs no function but allows the Go object to collect
	// information on the new thread.
	// - label is an optional name that can be assigned to a Go goroutine thread
	Register(label ...string) (g Go)
	// AddError emits a non-fatal errors
	AddError(err error)
	// Go returns a Go object to be provided as a go-statement function-argument
	//	in a function call invocation launching a new gorotuine thread.
	//	- the new thread belongs to the same GoGroup thread-group as the Go
	//		object whose Go method was invoked.
	Go() (g Go)
	// SubGo returns a GoGroup thread-group whose fatal and non-fatel errors go to
	//	the Go object’s parent thread-group.
	//	- a SubGo is used to ensure sub-threads exiting prior to their parent thread
	//		or to facilitate separate cancelation of the threads in the subordinate thread-group.
	//	- fatal errors from SubGo threads are handled in the same way as those of the
	//		Go object, typically terminating the application.
	//   - the SubGo thread-group terminates when both its own threads have exited and
	//	- the threads of its subordinate thread-groups.
	SubGo(onFirstFatal ...GoFatalCallback) (subGo SubGo)
	// SubGroup returns a thread-group with its own error channel.
	//	- a SubGroup is used for threads whose fatal errors should be handled
	//		in the Go thread.
	//	- The threads of the Subgroup can be canceled separately.
	//		- SubGroup’s error channel collects fatal thread terminations
	//		- the SubGroup’s error channel needs to be read in real-time or after
	//		SubGroup termination
	//   - non-fatal errors in SubGroup threads are sent to the Go object’s parent
	//		similar to the AddError method
	//   - the SubGroup thread-group terminates when both its own threads have exited and
	//	- the threads of its subordinate thread-groups.
	SubGroup(onFirstFatal ...GoFatalCallback) (subGroup SubGroup)
	// Done indicates that this goroutine is exiting
	//	- err nil: successful exit
	//	- err non-nil: fatal error exit
	//	- —
	// 	- deferrable
	//   - Done makes a goroutine:
	//   - — awaitable and
	//   - — able to return error
	//   - — other needs of a goroutine is to initiate and detect cancel and
	//		submit non-fatal errors
	//	- Done(errp *error)
	Doner
	// Wait awaits exit of this Go thread
	Wait()
	// ch closes upon exit of this Go thread
	WaitCh() (ch AwaitableCh)
	// Cancel signals for the threads in this Go thread’s parent GoGroup thread-group
	// and any subordinate thread-groups to exit.
	Cancel()
	// Context will Cancel when the parent thread-group Cancels
	//	or Cancel is invoked on this Go object.
	//	- Subordinate thread-groups do not Cancel the context of the Go thread.
	Context() (ctx context.Context)
	// ThreadInfo returns thread data that is partially or fully populated
	//	- ThreadID may be invalid: threadID.IsValid.
	//	- goFunction may be zero-value: goFunction.IsSet
	//	- those values present after public methods of parl.Go has been invoked by
	//		the new goroutine
	ThreadInfo() (threadData ThreadData)
	// values always present
	Creator() (threadID ThreadID, createLocation *pruntime.CodeLocation)
	//	- ThreadID may be invalid: threadID.IsValid.
	//	- goFunction may be zero-value: goFunction.IsSet
	//	- those values present after public methods of parl.Go has been invoked by
	//		the new goroutine
	GoRoutine() (threadID ThreadID, goFunction *pruntime.CodeLocation)
	// GoID efficiently returns the goroutine ID that may be invalid
	//	- valid after public methods of parl.Go has been invoked by
	//		the new goroutine
	GoID() (threadID ThreadID)
	// EntityID returns a value unique for this Go
	//	- ordered: usable as map key or for sorting
	//	- always valid, has .String method
	EntityID() (goEntityID GoEntityID)
	fmt.Stringer
}

// GoFatalCallback receives the thread-group on its first fatal thread-exit
//   - GoFatalCallback is an optional onFirstFatal argument to
//   - — NewGoGroup
//   - — SubGo
//   - — SubGroup
type GoFatalCallback func(goGen GoGen)

// GoGen allows for new Go threads, new SubGo and SubGroup thread-groups and
// cancel of threads in the thread-group and its subordinate thread-groups.
//   - GoGen is value from NewGoGroup GoGroup SubGo SubGroup Go,
//     ie. any Go-interface object
type GoGen interface {
	// Go returns a Go object to be provided as a go statement function argument.
	Go() (g Go)
	// SubGo returns a thread-group whose fatal errors go to GoGen’s parent.
	//   - both non-fatal and fatal errors in SubGo threads are sent to GoGen’s parent
	// 		like Go.AddError and Go.Done.
	//		- therefore, when a SubGo thread fails, the application will typically exit.
	//		- by awaiting SubGo, Go can delay its exit until SubGo has terminated
	//   - the SubGo thread-group terminates when the its thread exits
	SubGo(onFirstFatal ...GoFatalCallback) (subGo SubGo)
	// SubGroup creates a sub-ordinate thread-group
	SubGroup(onFirstFatal ...GoFatalCallback) (subGroup SubGroup)
	// Cancel terminates the threads in the Go consumer thread-group.
	Cancel()
	// Context will Cancel when the parent thread-group Cancels.
	// Subordinate thread-groups do not Cancel this context.
	Context() (ctx context.Context)
}

// Thread Group interfaces and Factory

// GoGroup manages a hierarchy of threads
//   - GoGroup only terminates when:
//   - — the last thread in its hierarchy exits
//   - — [GoGroup.EnableTermination] is set to true when
//     no Go threads exist in the hierarchy
//   - the GoGroup hierarchy consists of:
//   - — managed goroutines returned by [GoGroup.Go]
//   - — a [SubGo] subordinate thread-group hierarchy returned by [GoGroup.SubGo]
//     that allows for a group of threads to be canceled or waited upon separately
//   - — a [SubGroup] subordinate thread-group hierarchy returned by [GoGroup.SubGroup]
//     that allows for a group of threads to exit with fatal errors without
//     canceling the GoGroup and for those threads to be
//     canceled or waited upon separately
//   - — each subordinate Go thread or SubGo or SubGroup subordinate thread-groups
//     can create additional threads and subordinate thread-groups.
//   - [GoGroup.Context] returns a context that is the context or parent context
//     of all the Go threads, SubGo and SubGroup subordinate thread-groups
//     in its hierarchy
//   - [GoGroup.Cancel] cancels the GoGroup Context,
//     thereby signaling to all threads in the GoGroup hierarchy to exit.
//     This will eventually terminate the GoGroup
//   - providing a parent context to the GoGroup implementation allows
//     for terminating the GoGroup via this parent context
//   - A thread invoking [Go.Cancel] will signal to all threads in its
//     GoGroup or SubGo or SubGroup thread-groups to exit.
//     It will also signal to all threads in its subordinate thread-groups to exit.
//     This will eventually terminate its threadgroup and all that threadgroup’s
//     subordinate threadgroups.
//   - Alternatives to [parl.Go] is [parl.NewGoResult] and [parl.NewGoResult2]
//     that only provides being awaitable to a goroutine
//   - —
//   - from this thread-group will terminate all threads in this
//     and subordinate thread-groups if this thread-group was provided
//     the FirstFailTerminates option, which is default.
//   - A fatal thread-termination in a sub thread-group only affects this
//     thread-group if the sub thread-group was provided a nil fatal function,
//     the FirstFailTerminates option, which is default, and no explicit
//     FailChannel option.
//   - Fatal thread terminations will propagate to parent thread-groups if
//     this thread group did not have a fatal function provided and was not
//     explicitly provided the FailChannel option.
//   - A Cancel in this thread-group or in a parent context cancels threads in
//     this and all subordinate thread-groups.
//   - A Cancel in a subordinate thread-group does not affect this thread-group.
//   - Wait in this thread-group wait for threads in this and all subordinate
//     thread-groups.
type GoGroup interface {
	// Go returns a Go object to be provided as a go statement function argument.
	Go() (g Go)
	// SubGo returns a thread-group whose fatal errors go to Go’s parent.
	//   - both non-fatal and fatal errors in SubGo threads are sent to Go’s parent
	// 		like Go.AddError and Go.Done.
	//		- therefore, when a SubGo thread fails, the application will typically exit.
	//		- by awaiting SubGo, Go can delay its exit until SubGo has terminated
	//   - the SubGo thread-group terminates when the its thread exits
	SubGo(onFirstFatal ...GoFatalCallback) (subGo SubGo)
	// SubGroup creates a sub-ordinate GoGroup.
	//	- SubGroup fatal and non-fatal errors are sent to the parent GoGroup.
	//	-	SubGroup-context initiated Cancel only affect threads in the SubGroup thread-group
	//	- parent-initiated Cancel terminates SubGroup threads
	//	- SubGroup only awaits SubGroup threads
	//	- parent await also awaits SubGroup threads
	SubGroup(onFirstFatal ...GoFatalCallback) (subGroup SubGroup)
	// GoError returns a channel returning fatal and not-fatal errors and thread exits.
	//	- see GoError for exact error categories
	//
	// Usage:
	//
	//	var goErrors = goGroup.GoError()
	//	for goError := goErrors.Init(); goErrors.Condition(&goError); {
	//	  …
	//	var goEndCh = goErrors.EmptyCh()
	//	for {
	//	  select {
	//	  case <-goEndCh:
	//	    goEndCh = nil
	//	    …
	//	  case <-goErrors.DataWaitCh():
	//	    var goError, hasValue = goErrors.Get()
	//	    if !hasValue {
	//	      continue
	//	    …
	GoError() (goErrorSource IterableAllSource[GoError])
	// Wait waits for this thread-group to end
	Wait()
	// ch closes upon exit of this Go thread
	WaitCh() (ch AwaitableCh)
	// EnableTermination allows to temporarily prevent the GoGroup from
	// terminating.
	//   - allowTermination: desired state [AllowTermination] [PreventTermination]
	//   - mayTerminate: if allowTermination missing, returns current state.
	//     Otherwise, returns the set state
	//   - — a terminated go-group always returns AllowTermination
	//	- —
	//	- EnableTermination is initially [AllowTermination]
	//   - invoking with AllowTermination while child-object count is zero,
	//     terminates the thread-group regardless of previous enableTermination state.
	//     This is used prior to Wait when a thread-group was not used.
	//     Using the alternative Cancel would signal to threads to exit.
	//	- invoked with [PreventTermination] allows for the number of managed
	//		threads to be temporarily zero without terminating the thread-group.
	//		For example, if exiting threads during creation of multiple threads causes
	//		thread-count to reach zero, the thread-group prematurely terminates
	EnableTermination(allowTermination ...GoTermination) (mayTerminate GoTermination)
	// Cancel terminates the threads in this and subordinate thread-groups.
	Cancel()
	// Context will Cancel when the parent context Cancels.
	// Subordinate thread-groups do not Cancel this context.
	Context() (ctx context.Context)
	// the available data for all threads
	Threads() (threads []ThreadData)
	// threads that have been named ordered by name
	NamedThreads() (threads []ThreadData)
	// SetDebug enables debug logging on this particular instance
	//	- parl.NoDebug
	//	- parl.DebugPrint
	//	- parl.AggregateThread
	SetDebug(debug GoDebug, log ...PrintfFunc)
	fmt.Stringer
}

type SubGo interface {
	// Go returns a [Go] object managing a thread of the GoGroup thread-group
	// by providing the g value as a go-statement function argument.
	Go() (g Go)
	// SubGo returns a thread-group whose fatal errors go to Go’s parent.
	//   - both non-fatal and fatal errors in SubGo threads are sent to Go’s parent
	// 		like Go.AddError and Go.Done.
	//		- therefore, when a SubGo thread fails, the application will typically exit.
	//		- by awaiting SubGo, Go can delay its exit until SubGo has terminated
	//   - the SubGo thread-group terminates when the its thread exits
	SubGo(onFirstFatal ...GoFatalCallback) (subGo SubGo)
	// SubGroup creates a sub-ordinate GoGroup.
	//	- SubGroup fatal and non-fatal errors are sent to the parent GoGroup.
	//	-	SubGroup-context initiated Cancel only affect threads in the SubGroup thread-group
	//	- parent-initiated Cancel terminates SubGroup threads
	//	- SubGroup only awaits SubGroup threads
	//	- parent await also awaits SubGroup threads
	SubGroup(onFirstFatal ...GoFatalCallback) (subGroup SubGroup)
	// Wait waits for all threads of this thread-group to terminate.
	Wait()
	// returns a channel that closes on subGo end similar to Wait
	WaitCh() (ch AwaitableCh)
	// EnableTermination allows to temporarily prevent the GoGroup from
	// terminating.
	//   - allowTermination: desired state [AllowTermination] [PreventTermination]
	//   - mayTerminate: if allowTermination missing, returns current state.
	//     Otherwise, returns the set state
	//   - — a terminated go-group always returns AllowTermination
	//	- —
	//	- EnableTermination is initially [AllowTermination]
	//   - invoking with AllowTermination while child-object count is zero,
	//     terminates the thread-group regardless of previous enableTermination state.
	//     This is used prior to Wait when a thread-group was not used.
	//     Using the alternative Cancel would signal to threads to exit.
	//	- invoked with [PreventTermination] allows for the number of managed
	//		threads to be temporarily zero without terminating the thread-group.
	//		For example, if exiting threads during creation of multiple threads causes
	//		thread-count to reach zero, the thread-group prematurely terminates
	EnableTermination(allowTermination ...GoTermination) (mayTerminate GoTermination)
	// Cancel terminates the threads in this and subordinate thread-groups.
	Cancel()
	// Context will Cancel when the parent context Cancels.
	// Subordinate thread-groups do not Cancel this context.
	Context() (ctx context.Context)
	// the available data for all threads
	Threads() (threads []ThreadData)
	// threads that have been named ordered by name
	NamedThreads() (threads []ThreadData)
	// SetDebug enables debug logging on this particular instance
	//   - parl.NoDebug
	//   - parl.DebugPrint
	//   - parl.AggregateThread
	SetDebug(debug GoDebug, log ...PrintfFunc)
	fmt.Stringer
}

type SubGroup interface {
	SubGo
	// GoError returns a channel returning fatal and not-fatal errors and thread exits.
	//	- see GoError for exact error categories
	//
	// Usage:
	//
	//	var goErrors = goGroup.GoError()
	//	for goError := goErrors.Init(); goErrors.Condition(&goError); {
	//	  …
	//	var goEndCh = goErrors.EmptyCh()
	//	for {
	//	  select {
	//	  case <-goEndCh:
	//	    goEndCh = nil
	//	    …
	//	  case <-goErrors.DataWaitCh():
	//	    var goError, hasValue = goErrors.Get()
	//	    if !hasValue {
	//	      continue
	//	    …
	GoError() (goErrorSource IterableAllSource[GoError])
	// FirstFatal allows to await or inspect the first thread terminating with error.
	// it is valid if this SubGo has LocalSubGo or LocalChannel options.
	// To wait for first fatal error using multiple-semaphore mechanic:
	//  firstFatal := g0.FirstFatal()
	//  for {
	//    select {
	//    case <-firstFatal.Ch():
	//    …
	// To inspect first fatal:
	//  if firstFatal.DidOccur() …
	FirstFatal() (firstFatal *OnceWaiterRO)
}

type GoFactory interface {
	// NewGo returns a light-weight thread-group.
	//	- GoGroup only receives Cancel from ctx, it does not cancel this context.
	NewGoGroup(ctx context.Context, onFirstFatal ...GoFatalCallback) (g0 GoGroup)
}

const (
	// [GoGroup] [SubGo] [SubGroup] may again terminate
	AllowTermination GoTermination = iota
	// [GoGroup] [SubGo] [SubGroup] will not terminate until [AllowTermination]
	//	- used to avoid exiting threads causing thread-count to reach zero
	//		during thread-creation
	PreventTermination
)

// data types

// GoError is an error or a thread exit associated with a goroutine
//   - GoError encapsulates the original unadulterated error
//   - GoError provides context for taking action on the error
//   - Go provides the thread associated with the error. All GoErrors are associated with
//     a Go object
//   - because GoError is both error and fmt.Stringer, to print its string representation
//     requires using the String() method, otherwise fmt.Printf defaults to the Error()
//     method
type GoError interface {
	error // Error() string
	// Err retrieves the original error value
	Err() (err error)
	// ErrString returns string representation of error
	//   - if no error “OK”
	//   - if not debug or panic, short error with location
	//   - otherwise error with stack trace
	ErrString() (errString string)
	// Time provides when this error occurred
	Time() time.Time
	// IsThreadExit determines if this error is a thread exit, ie. GeExit GePreDoneExit
	//	- thread exits may have err nil
	//	- fatals are non-nil thread exits that may require specific actions such as
	//		application termination
	IsThreadExit() (isThreadExit bool)
	// IsFatal determines if this error is a fatal thread-exit, ie. a thread exiting with non-nil error
	IsFatal() (isThreadExit bool)
	// ErrContext returns in what situation this error occurred
	ErrContext() (errContext GoErrorContext)
	// Go provides the thread and goroutine emitting this error
	Go() (g0 Go)
	fmt.Stringer
}

// ThreadData is information about a Go object thread.
//   - initially, only Create is present
//   - Name is only present for threads that have been named
type ThreadData interface {
	// threadID is the ID of the running thread assigned by the go runtime
	//	- IsValid method checks if value is present
	//	- zero value is empty string
	//	- .ThreadID().String(): "5"
	ThreadID() (threadID ThreadID)
	// createLocation is the code line of the go statement function-call
	// creating the goroutine thread
	// - IsSet method checks if value is present
	//	- Create().Short(): "g0.(*SomeType).SomeCode()-thread-data_test.go:73"
	Create() (createLocation *pruntime.CodeLocation)
	// Func returns the code line of the function of the running thread.
	// - IsSet method checks if value is present
	//	- .Func().Short(): "g0.(*SomeType).SomeFunction()-thread-data_test.go:80"
	Func() (funcLocation *pruntime.CodeLocation)
	// optional thread-name assigned by consumer
	//	- zero-value: empty string "" for threads that have not been named
	Name() (label string)
	// Short returns a short description of the thread "label:threadID" or fmt.Stringer
	//	- "myThreadName:4"
	//	- zero-value: "[empty]" ThreadDataEmpty
	//	- nil value: "threadData:nil" ThreadDataNil
	Short() (short string)
	// all non-empty fields: [label]:[threadID]_func:[Func]_cre:[Create]
	//	- "myThreadName:5_func:g0.(*SomeType).SomeFunction()-thread-data_test.go:80_cre:g0.(*SomeType).SomeCode()-thread-data_test.go:73"
	//	- zero-value: "[empty]" ThreadDataEmpty
	fmt.Stringer
}

const (
	// GeNonFatal indicates a non-fatal error ocurring during processing.
	// err is non-nil
	GeNonFatal GoErrorContext = iota + 1
	// GePreDoneExit indicates an exit value of a subordinate goroutine,
	// other than the final exit of the last running goroutine.
	// err may be nil
	GePreDoneExit
	// A SubGroup with its own error channel is sending a
	// locally fatal error not intended to terminate the app
	GeLocalChan
	// A thread is requesting app termination without a fatal error.
	//	- this could be a callback
	GeTerminate
	// GeExit indicates exit of the last goroutine.
	// err may be nil.
	// The error channel may close after GeExit.
	GeExit
)

const (
	// default, no debug printing
	NoDebug GoDebug = iota
	// debug print on go-group events
	DebugPrint
	// the go-group internally collects extended thread information
	AggregateThread
)

// Iterator is an for range iterator over T
type GoErrorIterator interface {
	// Seq is an iterator over sequences of individual values.
	// When called as seq(yield), seq calls yield(v) for
	// each value v in the sequence, stopping early if yield returns false.
	GoError(yield func(value GoError) (keepGoing bool))
}

// GoErrorIterator.GoError is iter.Seq
//   - type Seq[V any] func(yield func(V) bool)
var _ = func(i GoErrorIterator) (seq iter.Seq[GoError]) { return i.GoError }
