/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package parl handles inter-thread communication and controls parallelism

Logging of extected output is via Out(string, ...interface{}).
Parl logging uses comma separator for numbers and is thread safe

Log(string, ...interface{}) always outputs to stderr.
Console is the same intended to be used for command-line interactivity.
SetDebug(true) appends code location

Info is active by default and outputs to stderr.
SetSilent(true) removes this output.
SetDebug(true) appends code location
IsSilent deteremines if Info printing applies

Debug only prints if SetDebug(true) or the code location matches SetInfoRegexp().
The string matched for regular expression looks like: “github.com/haraldrudell/parl.FuncName”
IsThisDebug determines if debug is active for the executing function

parl.D is intended for temporary printouts to be removed before check-in

parl provides generic recovery for goroutines and functions:
capturing panics, annotating and storing errors and invoking an error handling function on errors:

  func f() (err error) {
		defer parl.Recover(parl.Annotation(), &err, onError func(e error) { … })
	…
Default error string: “Recover from panic in somePackage.someFunction: 'File not found'.
For multiple errors, Recover uses error116 error lists,
while Recover2 instead invokes onError multiple times

Parl is about 9,000 lines of Go code with first line written on November 21, 2018

On 3/16/2022 Parl was open-sourced under an ISC License

© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
*/
package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	Rfc3339s   = "2006-01-02 15:04:05-07:00"
	Rfc3339ms  = "2006-01-02 15:04:05.999-07:00"
	Rfc3339us  = "2006-01-02 15:04:05.999999-07:00"
	Rfc3339ns  = "2006-01-02 15:04:05.999999999-07:00"
	Rfc3339sz  = "2006-01-02T15:04:05Z"
	Rfc3339msz = "2006-01-02T15:04:05.999Z"
	Rfc3339usz = "2006-01-02T15:04:05.999999Z"
	Rfc3339nsz = "2006-01-02T15:04:05.999999999Z"
)

func Errorf(format string, a ...interface{}) (err error) {
	return perrors.Errorf(format, a...)
}

func New(s string) error {
	return perrors.New(s)
}

var parlSprintf = message.NewPrinter(language.English).Sprintf

func Sprintf(format string, a ...interface{}) string {
	return parlSprintf(format, a...)
}

type Password interface {
	HasPassword() (hasPassword bool)
	Password() (password string)
}

type FSLocation interface {
	Directory() (directory string)
}
