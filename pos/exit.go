/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"errors"
	"fmt"
	"os"
)

const (
	statusCode      = 1
	StatusCodeUsage = 2
)

// Exit0 terminates the process successfully
func Exit0() {
	OsExit(0)
}

// Exit1 terminate the command echoing a failure to stderr and returning status code 1
func Exit1(err error) {
	Exit(statusCode, err)
}

// Exit1OneLine terminate the command echoing to stderr returning status code 1
func Exit1OneLine(err error) {
	if err == nil {
		err = errors.New("Exit1OneLine with err nil")
	}
	fmt.Fprintf(os.Stderr, "%v\n", err)
	OsExit(statusCode)
}

// Exit terminate the command echoing to stderr returning status code
func Exit(code int, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
	}
	OsExit(code)
}

// OsExit does os.Exit
func OsExit(code int) {
	os.Exit(code) // prints "exit status 1" to stderr
}
