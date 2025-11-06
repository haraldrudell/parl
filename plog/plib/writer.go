/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plib

import (
	"io"

	"github.com/haraldrudell/parl"
)

// Writer implements an [io.Writer] logging to log
type Writer struct {
	log parl.PrintfFunc
}

// Writer is [io.Writer]
var _ io.Writer = &Writer{}

func NewWriter(log parl.PrintfFunc) (writer *Writer) {
	if log == nil {
		log = parl.Log
	}
	return &Writer{log: log}
}

// Write converts write of bytes to log of string
func (w *Writer) Write(p []byte) (n int, err error) {
	n = len(p)
	w.log(string(p))
	return
}
