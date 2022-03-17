/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package ev provides standardized goroutine management

events contain thread completions, failures and any type of data items. A manager launches and controls a set of goroutines

ctx.Result is defered by a gorutine, captures panics and sends the goroutine’s result using Success or Failure

  func (ctx ev.Callee) {
    var err error
    defer func() {
      ctx.Result(&err, recover())
    }()
  …

© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
*/
package ev

import (
	"context"

	"github.com/google/uuid"
)

// EventRx receives events from goroutines
type EventRx <-chan Event

// GoID is a unique value identifying a goroutine
type GoID uuid.UUID

// Callee provides functions for a called goroutine
type Callee interface {
	context.Context                                // Deadline Done Err Value
	Success()                                      // Success indicates this goruote terminated successfully
	Failure(err error)                             // Failure indicates this goroutine failed with err
	Result(errp *error)                            // recovery for gorutines: usage: defer ctx.Result(&err)
	ResultV(errp *error, recoverValue interface{}) // recovery for gorutines
	Thread() (name string, ID GoID)                // Thread provides information about the thread
	Send(payload interface{})                      // Send allows a goroutine to send custom types
}

const (
	// TerminateAll terminates other goroutines on first termination
	TerminateAll CancelAction = iota
	// KeepGoing waits for all goroutines to complete even on errors
	KeepGoing
	// WhileOk terminates if any goroutine errors
	WhileOk
)

// CancelAction holds strategy for when a goruotine terminates
type CancelAction uint8

// Event is a message passed from goroutines
type Event interface {
	GoID() (gID GoID) // [16]byte
}

// Manager provides management of goroutines
type Manager interface {
	Cancel() (isEnd bool)                    // Cancel signals to all goroutines to terminate
	Events() (ch EventRx)                    // Events provides the channel emitting events from goroutines
	CalleeContext(ID ...string) (ctx Callee) // CalleeContext context for a new goroutine, 0 or 1 argument
	ProcessEvent(event Event) (err error)    // ProcessEvent sanity checks an EventRx event
	Action(evResult error, action CancelAction) (awaitThreads bool)
	IsEnd() (isEnd bool)                   // IsEnd returns true if all goroutines have terminated
	Threads() (names []string, IDs []GoID) // Threads provide information on currently managed goroutines
}
