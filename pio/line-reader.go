/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// LineReader reads a stream one line per Read invocation.
package pio

import (
	"errors"
	"io"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/slices"
)

const (
	newLine           = byte('\n')
	notFound          = -1
	defaultAllocation = 1024
	minBuffer         = 512
	maxLine           = 1024 * 1024
)

// LineReader reads a stream one line per Read invocation.
type LineReader struct {
	reader           io.Reader
	isEof            bool
	byts             []byte
	searchStartIndex int
	nextNewlineIndex int
}

func NewLineReader(reader io.Reader) (lineReader *LineReader) {
	if reader == nil {
		panic(perrors.NewPF("readre cannot be nil"))
	}
	return &LineReader{reader: reader, byts: []byte{}, nextNewlineIndex: notFound}
}

// Read returns a byte-sequence ending with newline if size of p is sufficient.
//   - if size of p is too short, the text will not end with newline
//   - if EOF without newline, text has no newline and err is io.EOF
func (rr *LineReader) Read(p []byte) (n int, err error) {

	for {

		// return a line if there is one or whatever fits p
		if len(rr.byts) > 0 {
			var index int
			var isEofLast bool
			if index = rr.nextNewlineIndex; index != notFound {
				rr.nextNewlineIndex = notFound
			} else if index = slices.Index(rr.byts[rr.searchStartIndex:], newLine); index != -1 {
				index += rr.searchStartIndex + 1 // include newline
				rr.searchStartIndex = index + 1
			} else if rr.isEof {
				index = len(rr.byts) // non-terminated lines prior to eof
				isEofLast = true
			}
			if index != -1 {
				if pLength := len(p); pLength < index {

					// part of rr.byts
					n = pLength
					rr.nextNewlineIndex = index - n // remember where the next end is
				} else {

					// all of rr.byts
					n = index
					if isEofLast {
						err = io.EOF
					}
				}
				copy(p, rr.byts[:n])
				pslices.TrimLeft(&rr.byts, n)
				rr.searchStartIndex -= n
				return // line found return: n >= 0 err == nil or EOF
			}
		}

		// return EOF if it is EOF
		if rr.isEof {
			err = io.EOF
			return // eof return: n == 0; err == io.EOF
		}

		// if rr.byts not empty, read into byts
		if len(rr.byts) > 0 {
			if n, err = rr.readToByts(len(p)); err != nil {
				return
			}
			continue
		}

		// read from rr.reader into p
		if n, err = rr.reader.Read(p); err != nil {
			if rr.isEof = errors.Is(err, io.EOF); rr.isEof {
				err = nil
			} else {
				return // rr.reader.Read error return
			}
		}
		if index := slices.Index(p[:n], newLine); index != -1 {
			index++ // include newline
			if index < n {

				// save text beyond newline in rr.byts
				rr.byts = append(rr.byts, p[index:n]...)
				n = index
			}
			return // full line in p return: n > 0, err == nil
		}

		// save to byts, then read more into byts
		rr.byts = append(rr.byts, p[:n]...)
		rr.searchStartIndex = n // start searching for newline index
	}
}

// ReadLine returns full lines, extending p as necessary
//   - len(line) is number of bytes
//   - max line length 1 MiB
//   - line will end with newLine unless 1 MiB or isEOF
//   - EOF is returned as isEOF true
func (rr *LineReader) ReadLine(p []byte) (line []byte, isEOF bool, err error) {

	// get line from p
	if capP := cap(p); capP == 0 {
		line = make([]byte, defaultAllocation)
	} else {
		line = p[:capP]
	}

	var n int
	defer func() {
		line = line[:n]
	}()
	for {

		// read appending to line
		var n0 int
		n0, err = rr.Read(line[n:])
		n += n0
		if err != nil {
			if isEOF = errors.Is(err, io.EOF); isEOF {
				err = nil // io.EOF is returned in isEOF, not in err
			}
			return // read error or EOF return
		} else if n0 > 0 && line[n-1] == newLine {
			return // full line return
		} else if cap(line) >= maxLine {
			return // 1 MiB line return
		} else if requiredLength := n + minBuffer; requiredLength > cap(line) {
			parl.D("requiredLength %d cap %d make %d", requiredLength, cap(line),
				requiredLength+(defaultAllocation-requiredLength%defaultAllocation)%defaultAllocation)
			newSlice := make([]byte, requiredLength+(defaultAllocation-requiredLength%defaultAllocation)%defaultAllocation)
			copy(newSlice, line[:n])
			line = newSlice
		}
	}
}

// readToByts returns n and err after reading into rr.byts
//   - lengthP is max number of bytes to be read
//   - rr.isEof is updated. io.EOF is not returned
func (rr *LineReader) readToByts(lengthP int) (n int, err error) {
	lengthByts := len(rr.byts)
	defer func() {
		parl.D("lengthByts %d cap(rr.byts): %d n: %d err: %s lengthP %d", lengthByts, cap(rr.byts), n, perrors.Short(err), lengthP)
		rr.byts = rr.byts[:lengthByts+n]
	}()

	requiredLength := lengthByts + lengthP
	if cap(rr.byts) < requiredLength {
		newSlice := make([]byte, requiredLength)
		copy(newSlice, rr.byts)
		rr.byts = newSlice
	} else {
		rr.byts = rr.byts[:requiredLength]
	}

	if n, err = rr.reader.Read(rr.byts[lengthByts:]); err != nil {
		if rr.isEof = errors.Is(err, io.EOF); rr.isEof {
			err = nil
		}
	}

	return
}
