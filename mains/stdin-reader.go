/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// StdinReader is a reader wrapping the unclosable [os.Stdin.Read]
//   - on error, the error is sent to addError and EOF is returned
type StdinReader struct {
	// option error submitting function
	errorSink parl.ErrorSink1
	// whether error has occured in [StdinReader.Read]
	isClosed atomic.Bool
	// optional value set to true on error
	isError *atomic.Bool
}

// NewStdinReader returns a error-free reader of standard input that closes on error
//   - errorSink receives any errors returned by [os.Stdin.Read] or runtime panic in this method.
//     If missing, errors are printed to stderr
//   - isError is an optional atomic set to true on first error
//   - [StdinReader.Read] returns bytes read from standard input until it closes.
//     The only error returned is io.EOF
func NewStdinReader(errorSink parl.ErrorSink1, isError *atomic.Bool) (stdinReader io.Reader) {
	return &StdinReader{
		errorSink: errorSink,
		isError:   isError,
	}
}

// Read reads from standard input
//   - on [os.Stdin.Read] error or panic, the reader closes
//   - n: the number of bytes read
//   - err: Read never returns any other error than [io.EOF]
//   - — on any error or panic, Read returns io.EOF
//   - errors and runtime panics are sent to the errorSink or printed to stderr
//   - [os.Stdin] cannot be closed so a blocking [StdinReader.Read] cannot be canceled
//   - if the stdin pipe is closed by another process,
//     Read keeps blocking but returns on the next keypress.
//     Then, an error os.ErrClosed is sent to the errorsink and io.EOF is returned
//   - on process exit, Read is unblocked as stdin is closed
func (r *StdinReader) Read(p []byte) (n int, err error) {

	// already closed case
	if r.isClosed.Load() {
		err = io.EOF
		return
	}

	var isPanic bool

	n, isPanic, err = r.read(p)

	// no error case
	if err == nil {
		return
	}

	// store error condition in object
	r.isClosed.Store(true)
	// if isError present, note error has occurred
	if r.isError != nil {
		r.isError.Store(true)
	}

	// do not submit or print EOF error
	//	- indication is r.isError true
	if err == io.EOF {
		return
	}

	// if another process closes stdin:
	// os.StdinRead error:
	// “read /dev/stdin: input/output error [*fs.PathError]
	// input/output error [syscall.Errno]”
	// isPanic: false

	// if addError present, submit error to it
	if r.errorSink != nil {
		err = perrors.ErrorfPF("os.Stdin.Read error: “%w” isPanic: %t",
			err, isPanic,
		)
		r.errorSink.AddError(err)
		err = io.EOF
		return
	}

	fmt.Fprintf(os.Stderr, "os.Stdin.Read error: “%s” isPanic: %t",
		perrors.Long(err),
		isPanic,
	)
	err = io.EOF

	return
}

// read invokes [os.Stdin.Read] capturing panic
func (r *StdinReader) read(p []byte) (n int, isPanic bool, err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err, &isPanic)

	n, err = os.Stdin.Read(p)

	return
}
