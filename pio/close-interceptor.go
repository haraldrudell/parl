/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"

	"github.com/haraldrudell/parl"
)

// CloseInterceptor is an interceptor for io types implementing Close:
//   - [io.Closer] [io.ReadCloser] [io.WriteCloser] [io.ReadWriteCloser]
//   - [io.ReadSeekCloser]
type CloseInterceptor interface {
	// CloseIntercept decides if Close should be invoked on the wrapped type
	//	- closer: the wrapped value as [io.Closer]: nil if Close is not available
	//	- label: name assigned to the wrapped io type
	//	- closeNo: the number for the Close invocation 1…
	//	- err non-nil: ignore doClose and instead return the error
	//	- underlying Close is a noop if the wrapped type does not implement Close
	//	- CloseIntercept is intended to:
	//	- — implement idempotent Close
	//	- — print stack traces for Close invocations
	//	- — prevent Close even if the wrapped object does implement it
	//	- — otherwise modify the behavior of Close
	//	- — CloseIntercept is invoked behind wrapper-specific lock.
	//		Only one invocation per closer is active at any one time
	CloseIntercept(closer io.Closer, label string, closeNo int) (err error)
}

var (
	// CloseStackTraces prints a stack trace for each close invocation
	//   - also makes Close idempotent
	CloseStackTraces CloseInterceptor = &stackTraceClose{}

	// CloseIdempotent makes close Idempotent
	//   - only the first Close invocation may receive any error
	CloseIdempotent CloseInterceptor = &idempotentClose{}
)

type (
	stackTraceClose struct{}
	idempotentClose struct{}
)

// print stack trace to standard error for each invocation
//   - Close is idempotent
func (c stackTraceClose) CloseIntercept(closer io.Closer, label string, closeNo int) (err error) {
	parl.Log("CloseIntercept %s no%d", label, closeNo)
	if closeNo > 1 || closer == nil {
		return
	}
	parl.Close(closer, &err)
	return
}

// make Close idempotent
func (c idempotentClose) CloseIntercept(closer io.Closer, label string, closeNo int) (err error) {
	if closeNo > 1 || closer == nil {
		return
	}
	parl.Close(closer, &err)
	return
}
