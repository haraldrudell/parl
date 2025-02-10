/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"io"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Printer offers error-free Print of string delegated to
// either [parl.PrintfFunc] or [io.Writer]
type Printer struct {
	// non-nil when delegating to eror-free string function
	printfFunc parl.PrintfFunc
	// non-nil when delegating to [io.Writer]
	writer io.Writer
}

// Print is error-free string printer to [parl.PrintfFunc] or [io.Writer]
func (p *Printer) Print(s string) {

	// check for empty string
	if s == "" {
		return
	}

	// printfFunc case
	if p.printfFunc != nil {
		p.printfFunc(s)
		return
	}

	// error case
	var err error
	if p.writer == nil {
		err = perrors.NewPF("both PrintfFunc and Writer nil")
		panic(err)
	}

	// writer case
	var byts = []byte(s)
	var totalBytes = len(byts)
	var bytesWritten int

	// write until all written
	var n int
	for bytesWritten < totalBytes {
		if n, err = p.writer.Write(byts[bytesWritten:]); perrors.IsPF(&err, "provided Writer.Write %w", err) {
			panic(err)
		}
		bytesWritten += n
	}
}

// SetPrintFunc sets the parl.PrintfFunc to use
func (p *Printer) SetPrintFunc(printfFunc parl.PrintfFunc) {
	p.printfFunc = printfFunc
	p.writer = nil
}

// SetWriter sets the [io.Writer] to use
func (p *Printer) SetWriter(writer io.Writer) {
	p.printfFunc = nil
	p.writer = writer
}
