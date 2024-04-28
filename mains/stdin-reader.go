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

// StdinReader is a reader wrapping the unclosable os.Stdin.Read
//   - on error, the error is sent to addError and EOF is returned
type StdinReader struct {
	addError parl.AddError
	isClosed atomic.Bool
	isError  *atomic.Bool
}

var _ io.Reader = &StdinReader{}

// NewStdinReader returns a reader that closes on error
func NewStdinReader(addError parl.AddError, isError *atomic.Bool) (reader *StdinReader) {
	return &StdinReader{
		addError: addError,
		isError:  isError,
	}
}

// Read reads from standard input
//   - on error, the reader closes
//   - errors are submitted separately or printed to stderr
func (r *StdinReader) Read(p []byte) (n int, err error) {
	if r.isClosed.Load() {
		err = io.EOF
		return
	}

	var isPanic bool

	n, isPanic, err = r.read(p)

	if err != nil {
		r.isClosed.Store(true)
		if r.isError != nil {
			r.isError.Store(true)
		}
		// os.StdinRead error:
		// “read /dev/stdin: input/output error [*fs.PathError]
		// input/output error [syscall.Errno]”
		// isPanic: false
		if r.addError != nil {
			err = perrors.ErrorfPF("os.Stdin.Read error: “%w” isPanic: %t",
				err, isPanic,
			)
			r.addError(err)
		} else {
			fmt.Fprintf(os.Stderr, "os.Stdin.Read error: “%s” isPanic: %t",
				perrors.Long(err),
				isPanic,
			)
		}
		err = io.EOF
	}

	return
}

// read invokes [os.Stdin.Read] capturing panic
func (r *StdinReader) read(p []byte) (n int, isPanic bool, err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err, &isPanic)

	n, err = os.Stdin.Read(p)

	return
}
