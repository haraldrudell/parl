/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

// Invoker allows to invoke a function returning error as a function with no return value. Thread Safe
//   - panics while executing FuncErr are recovered
//
// Thread-safe
type Invoker struct {
	funcErr func() (err error)

	invokeLock sync.RWMutex
	isDone     bool
	result     InvokeResult
}

type InvokeResult struct {
	IsPanic bool
	Err     error
}

// NewInvoker returns an object that can invoke a function returning error as a
// function with no return values. Thread-safe
func NewInvoker(funcErr func() (err error)) (invoker *Invoker) {
	return &Invoker{funcErr: funcErr}
}

// Func invokes the function returning error storing results in its fields
func (i *Invoker) Func() {
	isPanic, err := RecoverInvocationPanicErr(i.funcErr)
	i.invokeLock.Lock()
	defer i.invokeLock.RUnlock()

	i.isDone = true
	i.result.IsPanic = isPanic
	i.result.Err = err
}

func (i *Invoker) Result() (isDone, isPanic bool, err error) {
	i.invokeLock.RLock()
	defer i.invokeLock.RUnlock()

	isDone = i.isDone
	isPanic = i.result.IsPanic
	err = i.result.Err

	return
}

func (i *Invoker) InvokeResult() (invokeResult *InvokeResult) {
	i.invokeLock.RLock()
	defer i.invokeLock.RUnlock()

	var copy = i.result
	invokeResult = &copy

	return
}
