/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bytes"
	"io"
	"io/fs"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// CloserBuffer extends byte.Buffer to be [io.Closer]
type CloserBuffer struct {
	// Available() AvailableBuffer() Bytes() Cap() Grow() Len() Next()
	// Read() ReadByte() ReadBytes() ReadFrom() ReadRune() ReadString()
	// Reset() String() Truncate() UnreadByte() UnreadRune()
	// Write() WriteByte() WriteRune() WriteString() WriteTo()
	bytes.Buffer
	isClosed atomic.Bool
}

var _ io.Closer = &CloserBuffer{} // Close()

var _ io.Reader = &CloserBuffer{}     // Read()
var _ io.ReaderFrom = &CloserBuffer{} // ReadFrom()
var _ io.RuneReader = &CloserBuffer{} // ReadRune()
var _ io.ByteReader = &CloserBuffer{} // ReadByte()

var _ io.Writer = &CloserBuffer{}       // Write()
var _ io.WriterTo = &CloserBuffer{}     // WriteTo()
var _ io.StringWriter = &CloserBuffer{} // WriteString()
var _ io.ByteWriter = &CloserBuffer{}   // WriteByte()

var _ io.ByteScanner = &CloserBuffer{} // UnreadByte()
var _ io.RuneScanner = &CloserBuffer{} // UnreadRune()

// interfaces that do not exist
//var _ io.StringReader = &CloserBuffer{}
//var _ io.RuneWriter = &CloserBuffer{}
//var _ io.StringScanner = &CloserBuffer{}

// NewCloserBuffer returns an [bytes.Buffer] implementing [io.Closer]
//   - if buffer is present, it is copied
//   - implements:
//   - [io.Closer] [io.ReadCloser] [io.WriteCloser] [io.ReadWriteCloser]
//   - [io.ReadWriter]
//   - [io.Reader] [io.WriterTo]
//   - [io.Writer] [io.ReaderFrom]
//   - [io.ByteReader] [io.RuneReader]
//   - [io.StringWriter] [io.ByteWriter]
//   - [io.ByteScanner] [io.RuneScanner]
func NewCloserBuffer(buffer ...*bytes.Buffer) (closer *CloserBuffer) {
	c := CloserBuffer{}
	if len(buffer) > 0 {
		if b := buffer[0]; b != nil {
			c.Buffer = *b
		}
	}
	return &c
}

// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
func (b *CloserBuffer) Read(p []byte) (n int, err error) {
	if err = b.readCheck(); err != nil {
		return
	}
	return b.Buffer.Read(p)
}

// ReadByte reads and returns the next byte from the input or
// any error encountered. If ReadByte returns an error, no input
// byte was consumed, and the returned byte value is undefined.
func (b *CloserBuffer) ReadByte() (c byte, err error) {
	if err = b.readCheck(); err != nil {
		return
	}
	return b.Buffer.ReadByte()
}

// ReadFrom reads data from r until EOF or error.
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
func (b *CloserBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	if err = b.closeCheck(); err != nil {
		return
	}
	return b.Buffer.ReadFrom(r)
}

// ReadRune reads a single encoded Unicode character
// and returns the rune and its size in bytes. If no character is
// available, err will be set.
func (b *CloserBuffer) ReadRune() (r rune, size int, err error) {
	if err = b.readCheck(); err != nil {
		return
	}
	return b.Buffer.ReadRune()
}

// UnreadByte causes the next call to ReadByte to return the last byte read.
func (b *CloserBuffer) UnreadByte() (err error) {
	if err = b.readCheck(); err != nil {
		return
	}
	return b.Buffer.UnreadByte()
}

// UnreadRune causes the next call to ReadRune to return the last rune read.
func (b *CloserBuffer) UnreadRune() (err error) {
	if err = b.readCheck(); err != nil {
		return
	}
	return b.Buffer.UnreadRune()
}

// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
func (b *CloserBuffer) Write(p []byte) (n int, err error) {
	if err = b.closeCheck(); err != nil {
		return
	}
	return b.Buffer.Write(p)
}

// WriteByte writes a byte to w.
func (b *CloserBuffer) WriteByte(c byte) (err error) {
	if err = b.closeCheck(); err != nil {
		return
	}
	return b.Buffer.WriteByte(c)
}

// WriteString writes the contents of the string s to w, which accepts a slice of bytes.
func (b *CloserBuffer) WriteString(s string) (n int, err error) {
	if err = b.closeCheck(); err != nil {
		return
	}
	return b.Buffer.WriteString(s)
}

// WriteTo writes data to w until there's no more data to write or
// when an error occurs. The return value n is the number of bytes
// written. Any error encountered during the write is also returned.
func (b *CloserBuffer) WriteTo(w io.Writer) (n int64, err error) {
	if err = b.readCheck(); err != nil {
		return
	}
	return b.Buffer.WriteTo(w)
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (b *CloserBuffer) Reset() {
	b.isClosed.Store(false)
	b.Buffer.Reset()
}

// Close should only be invoked once.
// Close is not required for releasing resources.
func (b *CloserBuffer) Close() (err error) {
	if b.isClosed.Load() || !b.isClosed.CompareAndSwap(false, true) {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // already closed or other thread closed
	}
	// noop: there isn’t actually a close
	return
}

// readCheck check for close for any read method
//   - a closed stream has deferred close allowing to read until the end
func (b *CloserBuffer) readCheck() (err error) {
	if !b.isClosed.Load() || b.Buffer.Len() > 0 {
		return
	}
	err = perrors.ErrorfPF("%w", fs.ErrClosed)
	return
}

// closeCheck checks for close for writing methods
func (b *CloserBuffer) closeCheck() (err error) {
	if b.isClosed.Load() || !b.isClosed.CompareAndSwap(false, true) {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
	}
	return
}
