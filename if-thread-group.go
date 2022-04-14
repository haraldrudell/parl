/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

type ThreadGroup interface {
	Add()
	AddThread(input bool /* 220414 TODO ThreadInput*/, errorMethod ErrorMethod) (threadID ThreadID, thread Thread, threadControl ThreadControl)
	WaitForThread(threadID ThreadID)
	WaitPeriod(duration time.Duration) (isDone bool, threadList []ThreadControl)
	Wait()
}

type ThreadID string
type channelID string

const (
	// EMsharedChannel outputs thread errors in real-time using a channel shared by multiple threads
	EMsharedChannel ErrorMethod = iota + 1
	// EMdedicatedChannel outputs thread errors in real-time using a channle unique to the thread
	EMdedicatedChannel
	EMsharedParlErrors
	EMdedicatedParlErrors
)

type ErrorMethod uint8

type ThreadInput[T any] interface {
	Value() (value T)
}

type ThreadControl interface {
	SetErrorMethod()
	AddChannel(ID channelID)
}

type ThreadResult[T any] interface {
	Result() (result T)
}

type Thread interface {
	NotifyStart()
	Send(ID channelID, value interface{})
	Receive(ID channelID) (value interface{})
	AddError(err error)
	// Done signals thread possibly without error
	Done()
	// DoneFailure signals thread ending with error
	DoneFailure(err error)
}

type ThreadInputFactory[T any] interface {
	NewThreadInput(value T) (input ThreadInput[T])
}

type ThreadResultFactory[T any] interface {
	NewThreadResult(thread Thread, result T) ThreadResult[T]
}

type ThreadGroupFactory[T any] interface {
	NewThreadGroup() (threadGroup ThreadGroup)
}
