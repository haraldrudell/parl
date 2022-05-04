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

var v interface{}
var vFunc func(v interface{})
var done bool

// SetVFunc receives a function value from parl to receive the symbol value
func SetVFunc(fn func(v interface{})) {
	if done {
		return
	}
	done = v != nil
	if done {
		fn(v)
	} else {
		vFunc = fn
	}
}

// SetV receives a symbol value from a parl sub-package
func SetV(value interface{}) {
	if done {
		return
	}
	done = vFunc != nil
	if done {
		vFunc(value)
	} else {
		v = value
	}
}
