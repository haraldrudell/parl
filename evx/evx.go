/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package evx is deprecated by the github.com/haraldrudell/parl/g0 package 220429

Package evx contains declarations not essential to event handling
*/
package evx

// PrintLine is full-line text output from ev threads
type PrintLine string

// StatusText is used for ongoing progress print-outs
type StatusText string

// Warning provides errors that should not terminate the ev thread
type Warning struct {
	Err error
}
