/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package breakcycle

var short interface{}
var shortReceiver func(v interface{})
var g0PtimeDone bool

// ParlImport receives a function value from parl to receive the symbol value
func G0Import(receiver func(v interface{})) {
	if g0PtimeDone {
		return
	}
	g0PtimeDone = short != nil
	if g0PtimeDone {
		receiver(short)
	} else {
		shortReceiver = receiver
	}
}

// GoidExport receives a symbol value from a parl sub-package
func PtimeExport(shortValue interface{}) {
	if g0PtimeDone {
		return
	}
	g0PtimeDone = shortReceiver != nil
	if g0PtimeDone {
		shortReceiver(shortValue)
	} else {
		short = shortValue
	}
}
