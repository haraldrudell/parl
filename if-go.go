/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"fmt"
	"time"

	"github.com/haraldrudell/parl/pruntime"
)

// Thread interface

// Go provides methods for a running goroutione thread to be provided as a function
// argument in the go statement function call launching the thread.
//   - Go.Cancel affects this Go thread only.
//   - The Go Context is canceled when
//   - — the parent GoGroup thread-group’s context is Canceled or
//   - —a thread in the parent GoGroup thread-group initiates Cancel
//   - Cancel by threads in sub ordinate thread-groups do not Cancel this Go thread
type Go interface {
	// Register performs no function but allows the Go object to collect
	// information on the new thread.
	Register()
	// AddError emits a non-fatal error.
	AddError(err error)
	// Go returns a Go object to be provided as a go-statement function-argument
	//	in a function call invocation launching a new gorotuine thread.
	//	- the new thread belongs to the same GoGroup thread-group as the Go
	//		object whose Go method was invoked.
	Go() (g1 Go)
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
	// Done indicates that this goroutine has finished.
	//	- err == nil means successful exit
	//	- non-nil err indicates a fatal error.
	// 	- Done is deferrable.
	Done(errp *error)
	// Wait awaits exit of this Go thread.
	Wait()
	// CancelGo signals to this Go thread to exit.
	CancelGo()
	// Cancel signals for the threads in this Go thread’s parent GoGroup thread-group
	// and any subordinate thread-groups to exit.
	Cancel()
	// Context will Cancel when the parent thread-group Cancels
	//	or Cancel is invoked on this Go object.
	// Subordinate thread-groups do not Cancel the context of the Go thread.
	Context() (ctx context.Context)
	ThreadInfo() (threadID ThreadID, createLocation pruntime.CodeLocation,
		funcLocation pruntime.CodeLocation, isValid bool)
	fmt.Stringer
}

// GoFatalCallback receives the thread-group on its first fatal thread-exit
//   - GoFatalCallback is an optional onFirstFatal argument to
//   - —NewGoGroup
//   - — SubGo
//   - — SubGroup
type GoFatalCallback func(goGen GoGen)

// GoGen allows for new Go threads, new SubGo and SubGroup thread-groups and
// cancel of threads in the thread-group and its subordinate thread-groups.
//   - GoGen can be a GoGroup or a Go object
type GoGen interface {
	// Go returns a G1 object to be provided as a go statement function argument.
	Go() (g1 Go)
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

// GoGroup manages a thread-group.
//   - A thread from this thread-group will terminate all threads in this
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
	// Go returns a G1 object to be provided as a go statement function argument.
	Go() (g1 Go)
	// SubGo returns athread-group whose fatal errors go to G1’s parent.
	//   - both non-fatal and fatal errors in SubGo threads are sent to G1’s parent
	// 		like G1.AddError and G1.Done.
	//		- therefore, when a SUbGo thread fails, the application will typically exit.
	//		- by awaiting SubGo, G1 can delay its exit until SubGo has terminated
	//   - the SubGo thread-group terminates when the its thread exits
	SubGo(onFirstFatal ...GoFatalCallback) (subGo SubGo)
	// SubGroup creates a sub-ordinate G1Group.
	//	- SubGroup fatal and non-fatal errors are sent to the parent G1Group.
	//	-	SubGroup-context initiated Cancel only affect threads in the SubGroup thread-group
	//	- parent-initiated Cancel terminates SubGroup threads
	//	- SubGroup only awaits SubGroup threads
	//	- parent await also awaits SubGroup threads
	SubGroup(onFirstFatal ...GoFatalCallback) (subGroup SubGroup)
	// Ch returns a channel sending the all fatal termination errors when
	// the FailChannel option is present, or only the first when both
	// FailChannel and StoreSubsequentFail options are present.
	Ch() (ch <-chan GoError)
	// Wait waits for all threads of this thread-group to terminate.
	Wait()
	// Cancel terminates the threads in this and subordinate thread-groups.
	Cancel()
	// Context will Cancel when the parent context Cancels.
	// Subordinate thread-groups do not Cancel this context.
	Context() (ctx context.Context)
	fmt.Stringer
}

type SubGo interface {
	// Go returns a G1 object to be provided as a go statement function argument.
	Go() (g1 Go)
	// SubGo returns athread-group whose fatal errors go to G1’s parent.
	//   - both non-fatal and fatal errors in SubGo threads are sent to G1’s parent
	// 		like G1.AddError and G1.Done.
	//		- therefore, when a SUbGo thread fails, the application will typically exit.
	//		- by awaiting SubGo, G1 can delay its exit until SubGo has terminated
	//   - the SubGo thread-group terminates when the its thread exits
	SubGo(onFirstFatal ...GoFatalCallback) (subGo SubGo)
	// SubGroup creates a sub-ordinate G1Group.
	//	- SubGroup fatal and non-fatal errors are sent to the parent G1Group.
	//	-	SubGroup-context initiated Cancel only affect threads in the SubGroup thread-group
	//	- parent-initiated Cancel terminates SubGroup threads
	//	- SubGroup only awaits SubGroup threads
	//	- parent await also awaits SubGroup threads
	SubGroup(onFirstFatal ...GoFatalCallback) (subGroup SubGroup)
	// Wait waits for all threads of this thread-group to terminate.
	Wait()
	// Cancel terminates the threads in this and subordinate thread-groups.
	Cancel()
	// Context will Cancel when the parent context Cancels.
	// Subordinate thread-groups do not Cancel this context.
	Context() (ctx context.Context)
	fmt.Stringer
}

type SubGroup interface {
	SubGo
	// Ch returns a receive channel for fatal errors if this SubGo has LocalChannel option.
	Ch() (ch <-chan GoError)
	// FirstFatal allows to await or inspect the first thread terminating with error.
	// it is valid if this SubGo has LocalSubGo or LocalChannel options.
	// To wait for first fatal error using multiple-semaphore mechanic:
	//  firstFatal := g1.FirstFatal()
	//  for {
	//    select {
	//    case <-firstFatal.Ch():
	//    …
	// To inspect first fatal:
	//  if firstFatal.DidOccur() …
	FirstFatal() (firstFatal *OnceWaiterRO)
}

type GoFactory interface {
	// NewG1 returns a light-weight thread-group.
	//	- G1Group only receives Cancel from ctx, it does not cancel this context.
	NewGoGroup(ctx context.Context, onFirstFatal ...GoFatalCallback) (g1 GoGroup)
}

// data types

// GoError is an error or a thread exit associated with a goroutine
// Goer returns the Goer object handling the goroutine that originated the error
type GoError interface {
	error // Error() string
	// Err retrieves the original error value
	Err() (err error)
	// Time provides when this error occurred
	Time() time.Time
	// IsThreadExit determines if this error is a thread exit
	//	- thread exits may have err nil
	//	- fatals are non-nil thread exits that may require specific actions such as
	//		applicaiton termination
	IsThreadExit() (isThreadExit bool)
	// IsFatal determines if this error is a fatal thread-exit, ie. a thread exiting with non-nil error
	IsFatal() (isThreadExit bool)
	// ErrContext returns in what situation this error occurred
	ErrContext() (errContext GoErrorContext)
	// Go provides the thread and goroutine emitting this error
	Go() (g0 Go)
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
	GeLocalChan
	// GeExit indicates exit of the last goroutine.
	// err may be nil.
	// The error channel may close after GeExit.
	GeExit
)
