/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bufio"
	"errors"
	"io"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

// LineReader reads a [io.Reader] stream returing one line per Read invocation
//   - operates on efficient byte-level
//   - does not implement [io.WriteTo] or [io.Closer]
//   - when reading long lines, avoids dounle-copy
//   - alternatives:
//   - — [bufio.Scanner] limited to less than 64 KiB or set Buffer, default separator is “\n” or “\r\n”
//   - — [bufio.Reader.ReadString] limited to single-byte separator
//   - — [bufio.Reader.ReadLine] reads line segments, separator is “\n” or “\r\n”
type LineReader struct {
	// reader is the underlying blocking reader
	reader io.Reader
	// eof is non-nil if EOF error was received from reader
	eof error
	// b is buffer used for read invocations
	//	- 4 KiB
	b []byte
	// byts are bytes to be returned by line reader
	//	- on read from the underlying reader into b,
	//		byts is set to any bytes being read
	//	- lines and bytes are then copied to the buffer
	//		provided to line-reader Read
	//	- byts is a slice expression of b or nil
	byts []byte
	// needSearch is true when byts is non-empty
	// and have not been searched for newlines
	needSearch bool
	// afterNewLine is the index after any newline found in byts
	//	- 0 means no newline
	//	- 1 means first byte of bytes is newline
	//	- range is 0…len(byts)
	afterNewLine int
	// sliceList is a list of buffers read from reader
	sliceList [][]byte
	// number of bytes in sliceList
	sliceListElements int
	//	- need to know if byts has not been searched
	//	- — searchStartIndex 0 means no search was conducted
	//	- need to know if byts was partially searched
	//	- — searchStartIndex > 0 means a partial search was conducted
	//	- need to know if a failed search of entire byts was conducted:
	//	- — searchStartIndex == len(byts) means the buffer was scanned
	//	- need to know if a successful search of byts was conducted
	//	- — afterNewline 1… is the index after any found newline
	//	- — afterNewline 0 means no newline has been found
	// searchStartIndex is position in byts to search for newline
	//	- 0 means no search was conducted
	//	- 1…len(byts)-1 means searched this far
	//	- len(byts) means the buffer was scanned
	// searchStartIndex int
}

// NewLineReader reads a [io.Reader] stream returning one line per Read invocation
//   - reader: a blocking reader
//   - —
//   - [LineReader.Read] returns line fitted into a buffer with newline terminator
//   - [LineReader.ReadLine] returns lines up to 1 MiB with terminator
//   - line delimiter is ‘\n’
//   - LineReader operates on efficient bytes, not strings
//   - LineReader does not implement [io.WriteTo] or [io.Closer]
func NewLineReader(reader io.Reader) (r *LineReader) {
	parl.NilPanic("reader", reader)
	return &LineReader{
		reader: reader,
		b:      make([]byte, defaultAllocation),
	}
}

// Read returns a byte-sequence ending with newline separator if size of p is sufficient
//   - p: buffer for bytes
//   - n: number of bytes read
//   - err: EOF or error returned by reader
//   - —
//   - line delimiter is ‘\n’
//   - if size of p is too short, the text will not end with newline
//   - if EOF without newline, text has no newline and err is io.EOF
func (r *LineReader) Read(p []byte) (n int, err error) {

	// check for zero-length buffer
	if len(p) == 0 {
		return
	}

	var n0 int
	var isFullLine bool
	for {

		// if there are bytes in byts, copy to p
		if len(r.byts) > 0 {

			// n0 is number of moved bytes
			//	- byts is updated
			n0, isFullLine = r.moveLineOrBytes(p)

			// update n
			n += n0

			// done check
			if isFullLine {
				return // ending with newline return
			} else if n0 == len(p) {
				// entire p case
				return // p filled return
			}

			// skip from p
			p = p[n0:]
		}
		// byts is empty, p is non-empty

		// check for read to end
		if r.eof != nil {
			err = r.eof
			return // eof return: p may contain bytes but no newline
		}

		// read more bytes
		if n0, err = r.readToByts(); err != nil {
			return // read error return
		} else if n0 > 0 {
			continue // byts is non-empty
		}

		// empty read case: n0 == 0
		// zero-read, byts is empty
		if r.eof != nil {
			err = r.eof
		}
		return // zero-byte read return: p may contain bytes but no newline
	}
}

// ReadLine returns full lines, extending p as necessary
//   - p: optional buffer. If missing, any returned line is allocated
//   - line non-nil: returned bytes typically ending with newline
//   - — no ending newline if line longer than 1 MiB or EOF without terminator
//   - — line does not share underlying array with
//     any provided p if line does not fit p.
//     Allocation takes place up to 1 MiB
//   - — isEOF is true if stream ended without newline
//   - line nil: stream ended with newline-EOF or error:
//     either isEOF is true or err non-nil
//   - isEOF: true if underlying reader was read to EOF
//   - err: any error other than EOF returned by reader
//   - — EOF is returned as isEOF true
func (r *LineReader) ReadLine(p ...[]byte) (line []byte, isEOF bool, err error) {

	// as bytes are read from r.reader:
	//	- lines are likely short, tens of bytes
	//	- it is not known where newlines are,
	//		so speculative read is required
	//	- data can be read to p or to a buffer
	//	- any bytes read beyond a newline
	//		must be stored in buffer prior to return
	//	- therefore, a buffer must be allocated
	//	- therefore, it is the easiest to read to buffer

	// to avoid allocations, line should be set to p whenever possible

	var (
		// nextBuffer is remaining space in nextBuffer0
		//	-	full-length slice filled with bytes from byts
		//	- nextBuffer is a slice expression of nextBuffer0
		nextBuffer []byte
		// nextBuffer0 is the original full-length slice for nextBuffer
		nextBuffer0 []byte
		// n0 is temporary number of bytes read
		n0 int
		// isFullLine is true if entire line was read
		isFullLine bool
		// nextBufferN is the number of valid bytes in nextBuffer0
		nextBufferN int
	)
	if len(p) > 0 {
		if p0 := p[0]; len(p0) > 0 {
			nextBuffer0 = p0
			nextBuffer = p0
		}
	}

	for {

		// consume byts
		if len(r.byts) > 0 {

			// ensure nextBuffer has space
			if len(nextBuffer) == 0 {
				if len(nextBuffer0) > 0 {
					r.appendLine(nextBuffer0)
				}
				nextBuffer = r.allocateBuffer()
				nextBuffer0 = nextBuffer
				nextBufferN = 0
			}
			// byts is non-empty, nextBuffer is non-empty

			// move bytes or line to nextBuffer
			n0, isFullLine = r.moveLineOrBytes(nextBuffer)
			// n0 > 0
			nextBufferN += n0
			nextBuffer = nextBuffer[n0:]
			// byts and nextBuffer may be empty

			// check for full line
			if isFullLine || // read to newline
				r.sliceListElements+nextBufferN >= maxLine { // read to 1 MiB
				break // line in sliceList and nextBuffer0[:nextBufferN]
			} else if len(r.byts) > 0 {
				continue // consume remaining byts
			}
		}
		// r.byts is empty

		// need to read more bytes or get to EOF
		if r.eof != nil {
			break // read to eof
		}

		// read more bytes
		if n0, err = r.readToByts(); err != nil {
			return // read error return
		} else if n0 == 0 {
			break // zero-byte read
		} else if n0 < len(r.b) || // read less than 4 KiB
			r.afterNewLine > 0 || // buffer contains newline
			len(r.b) <= cap(nextBuffer0)-nextBufferN { // 4 KiB fits nextBuffer
			continue // read less than 4 KiB or found newline
		}

		// 4 KiB case
		//	- r.b contains 4 KiB without newline
		//	- instead of double-copy, move the buffer
		if nextBufferN > 0 {
			r.appendLine(nextBuffer0[:nextBufferN])
			nextBuffer = nil
			nextBuffer0 = nil
			nextBufferN = 0
		}
		r.appendLine(r.b)
		r.b = r.allocateBuffer()
		r.byts = nil
		if r.sliceListElements >= maxLine {
			break // 1 MiB line
		}
	}
	// read to end of line, to 1 MiB length or
	// encountered end of file or zero-byte read

	// set line
	if nextBuffer0 != nil && nextBufferN > 0 {
		nextBuffer = nextBuffer0[:nextBufferN]
	} else {
		nextBuffer = nil
	}
	line = r.getLine(nextBuffer)

	isEOF = len(r.byts) == 0 && r.eof != nil

	return
}

// moveLineOrBytes moves any line or bytes from byts to p
//   - p buffer to write bytes to
//   - n0 number of bytes copied to p
//   - isFullLine: true if moved bytes end with newline
//   - —
//   - removes the moved bytes from byts
func (r *LineReader) moveLineOrBytes(p []byte) (n0 int, isFullLine bool) {

	// ensure searched for newline
	if r.needSearch {
		r.search()
	}

	// copy bytes to p
	if r.afterNewLine == 0 {
		n0 = copy(p, r.byts)
	} else {
		n0 = copy(p, r.byts[:r.afterNewLine])
		isFullLine = n0 == r.afterNewLine
	}
	// n0 > 0

	// remove from byts
	if n0 == len(r.byts) {
		r.byts = nil
	} else {
		r.byts = r.byts[n0:]
	}

	// entire line case
	if isFullLine {
		r.afterNewLine = 0
		r.needSearch = r.byts != nil
		return // ending with newline return
	}

	// update afterNewLine
	if r.afterNewLine > 0 {
		r.afterNewLine -= n0
	}

	return // partial line return
}

// search looks for newline in byts
//   - needSearch should have been verified true
func (r *LineReader) search() {
	r.needSearch = false
	if index := slices.Index(r.byts, newLine); index != -1 {
		r.afterNewLine = index + 1
	}
}

// readToByts reads from r.reader into byts
//   - n: number of bytes read
//   - err: any error from Read other than io.EOF
//   - —
//   - byts must have been verified to be empty
//   - read bytes are in byts
//   - newline search is performed
func (r *LineReader) readToByts() (n int, err error) {

	// read more bytes
	n, err = r.reader.Read(r.b)

	// process error
	if err != nil {
		if !errors.Is(err, io.EOF) {
			err = perrors.ErrorfPF("read %w", err)
			return // read error return
		} else if r.eof == nil {
			r.eof = err
		}
		err = nil
	}

	// zero-length read
	if n == 0 {
		return // zero-length read return
	}

	// update byts
	r.byts = r.b[:n]
	r.search()

	return
}

// allocateBuffer allocates a buffer to known newline or
// length of byts
func (r *LineReader) allocateBuffer() (b []byte) {

	// ensure searched for newline
	if r.needSearch {
		r.search()
	}

	// get size
	var size int
	if r.afterNewLine > 0 {
		size = r.afterNewLine
	} else {
		size = len(r.byts)
	}

	b = make([]byte, size)

	return
}

// appendLine stores buffer in sliceList
func (r *LineReader) appendLine(buffer []byte) {

	// ensure allocated
	if cap(r.sliceList) == 0 {
		r.sliceList = make([][]byte, sliceAllocSize)
	}

	r.sliceList = append(r.sliceList, buffer)
	r.sliceListElements += len(buffer)
}

// getLine build line
func (r *LineReader) getLine(buffer []byte) (line []byte) {

	if len(r.sliceList) == 0 {
		line = buffer
		return
	}

	var size = len(buffer)
	for i := range len(r.sliceList) {
		size += len(r.sliceList[i])
	}
	line = make([]byte, size)
	var s = line

	for i := range len(r.sliceList) {
		var n = copy(s, r.sliceList[i])
		s = s[n:]
	}
	clear(r.sliceList)
	r.sliceList = r.sliceList[:0]

	copy(s, buffer)

	return
}

const (
	// newLine is the single-byte line terminator
	newLine = byte('\n')
	// default allocation of b: 4 KiB
	defaultAllocation = 4096
	// max length returned by ReadLine: 1 MiB
	maxLine = 1024 * 1024
	//	- number of 4 KiB buffers to 1 MiB rounded down
	//	- plus one for overage and rounding
	//	- plus one for short p provided
	//	- plus one for short byts on invocation
	sliceAllocSize = maxLine/defaultAllocation + 3
)

// func bufio.NewScanner(r io.Reader) *bufio.Scanner
//   - limited to 64 KiB or set buffer
//   - longer lines is error
//   - default separator is “\n” or “\r\n”
var _ = bufio.NewScanner

// func bufio.NewReader(rd io.Reader) *bufio.Reader
//   - 4 KiB buffer
//   - uses copy to move bytes avoiding allocation
var _ = bufio.NewReader

// func bufio.NewReaderSize(rd io.Reader, size int) *bufio.Reader
var _ = bufio.NewReaderSize

// func (b *bufio.Reader) ReadString(delim byte) (string, error)
//   - can only do byte separator
// var _ = (&bufio.Reader{}).ReadString

// func (b *bufio.Reader) ReadLine() (line []byte, isPrefix bool, err error)
//   - separator is “\n” or “\r\n”
//   - reads parts of line
// var _ = (&bufio.Reader{}).ReadLine
