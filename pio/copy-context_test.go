/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

func TestCopyContext(t *testing.T) {
	//t.Fail()

	var text = []byte("Hello, World")
	var noWrite = 0

	var written int64
	var err error
	// has ReadFrom and WriteTo but not [io.Closable]
	var reader io.Reader
	// has ReadFrom and WriteTo but not [io.Closable]
	var writer io.Writer
	var ctx context.Context
	var cancelFunc context.CancelFunc
	var noBuffer []byte

	// a copy that completes should not error
	//	- can use bytes.NewReader
	//	- can use bytes.Buffer, this is common with exec.Cmd
	reader = bytes.NewBuffer(text)
	writer = &bytes.Buffer{}
	ctx = context.Background()
	// uses WriterTo
	written, err = CopyContext(writer, reader, noBuffer, ctx)
	if err != nil {
		t.Errorf("copyComplete err: %s", perrors.Short(err))
	}
	if int(written) != len(text) {
		t.Errorf("copyComplete written: %d exp %d", written, len(text))
	}

	// cancel via thread should not write and instead return error
	ctx = context.Background()
	ctx = parl.AddNotifier1(ctx, func(stack parl.Stack) {
		t.Log("contextCancelListener")
		var _ parl.NotifierFunc
	})
	ctx, cancelFunc = context.WithCancel(ctx)
	reader = NewBReadCloser(
		bytes.NewBuffer(text),
		cancelFunc,
		t,
	)
	writer = NewBWriteCloser(&bytes.Buffer{}, t)
	t.Logf("context is 0x%x", parl.Uintptr(ctx))
	written, err = CopyContext(writer, reader, noBuffer, ctx)
	t.Logf("cancel CopyContext: written: %d err: %s", written, perrors.Short(err))
	if err == nil {
		t.Error("cancelThread missing error")
	} else if !errors.Is(err, context.Canceled) {
		t.Errorf("cancelThread bad err: %s", perrors.Long(err))
	}
	if int(written) != noWrite {
		t.Errorf("copyComplete written: %d exp %d", written, noWrite)
	}

}

// BReadCloser invokes cancelFunc on Read
//   - is [io.Closable] to get thread
//   - does not implement Write To
type BReadCloser struct {
	b          *bytes.Buffer
	cancelFunc context.CancelFunc
	t          *testing.T
}

// BReadCloser invokes cancelFunc on Read
//   - is [io.Closable] to get thread
//   - does not implement Write To
func NewBReadCloser(
	b *bytes.Buffer,
	cancelFunc context.CancelFunc,
	t *testing.T,
) (b2 *BReadCloser) {
	return &BReadCloser{
		b:          b,
		cancelFunc: cancelFunc,
		t:          t,
	}
}

// Read invokes cancelFunc
func (b *BReadCloser) Read(p []byte) (n int, err error) {
	b.t.Log("Read invoking cancelFunc")
	b.cancelFunc()
	return b.b.Read(p)
}

// Close makes [io.Closable]
func (b *BReadCloser) Close() (err error) { return }

// BWriteCloser logs if Write is invoked
//   - does not implement readFrom
type BWriteCloser struct {
	b *bytes.Buffer
	t *testing.T
}

// BWriteCloser logs if Write is invoked
//   - does not implement readFrom
func NewBWriteCloser(b *bytes.Buffer, t *testing.T) (b2 *BWriteCloser) {
	return &BWriteCloser{b: b, t: t}
}

// BWriteCloser logs if Write is invoked
func (b *BWriteCloser) Write(p []byte) (n int, err error) {
	b.t.Log("Write")
	return b.b.Write(p)
}

// Close makes [io.Closable]
func (b *BWriteCloser) Close() (err error) { return }
