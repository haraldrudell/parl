/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
)

const (
	// error during delegated Read
	PeRead PIOErrorSource = iota + 1
	// error during delegated Write
	PeWrite
	// error during delegated Close
	PeClose
	// error from reads [io.Writer]
	PeReads
	// error from writes [io.Writer]
	PeWrites
)

// error source for a data tap
//   - [PeRead] [PeWrite] [PeClose] [PeReads] [PeWrites]
type PIOErrorSource uint8

func (s PIOErrorSource) String() (s2 string) { return sourceSet.StringT(s) }

// ReadError indicates an
type PIOError struct {
	error
	source PIOErrorSource
}

func NewPioError(source PIOErrorSource, e error) (err error) {
	return &PIOError{error: perrors.Stack(e), source: source}
}

func (e *PIOError) Error() (msg string) {
	if e.error != nil {
		msg = e.error.Error()
	}
	if e.source != 0 {
		if msg == "" {
			msg = e.source.String()
		} else {
			msg = e.source.String() + ": " + msg
		}
	}
	if msg == "" {
		msg = "uninitialized error"
	}
	msg = "tap: " + msg
	return
}

func (e *PIOError) Unwrap() (err error) { return e.error }

func (e *PIOError) PIOErrorSource() (source PIOErrorSource) { return e.source }

var sourceSet = sets.NewSet[PIOErrorSource]([]sets.SetElement[PIOErrorSource]{
	{ValueV: PeRead, Name: "error during delegated Read"},
	{ValueV: PeWrite, Name: "error during delegated Write"},
	{ValueV: PeClose, Name: "error during delegated Close"},
	{ValueV: PeReads, Name: "error from reads io.Writer"},
	{ValueV: PeWrites, Name: "error from writes io.Writer"},
})
