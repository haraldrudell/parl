/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package breakcycle breaks import cycles with parl

Typically import cycles appears with parl when:
— a sub-package imports symbols from parl and
— parl imports symbols from the sub-package

The import cycle can be broken by introducing a third package breakcycle.
— For breakcycle to be used it has to be imported
— Sub-package is allowed to import from parl and from breakcycle
— parl is allowed to import from breakcycle but not from sub-package
— breakcycle cannot import from parl or sub-package

sub→parl→bc←sub

— breakcycle initializes first because it is imported by both parl and sub-package
— parl is initialized next because it is imported by sub-package
— sub-package initializes last
In the time between parl and sub-package initialization,
— parl symbol values are invalid
*/
package breakcycle

var newStack interface{}
var newStackReceiver func(v interface{})
var goidParlDone bool

// ParlImport receives a function value from parl to receive the symbol value
func ParlImport(receiver func(v interface{})) {
	if goidParlDone {
		return
	}
	goidParlDone = newStack != nil
	if goidParlDone {
		receiver(newStack)
	} else {
		newStackReceiver = receiver
	}
}

// GoidExport receives a symbol value from a parl sub-package
func GoidExport(newStackValue interface{}) {
	if goidParlDone {
		return
	}
	goidParlDone = newStackReceiver != nil
	if goidParlDone {
		newStackReceiver(newStackValue)
	} else {
		newStack = newStackValue
	}
}
