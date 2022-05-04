/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parlos provides simplified functions related to the os package
package pos

import (
	"os"
	"strings"

	"github.com/haraldrudell/parl"
)

// ShortHostname gets hostname without domain part
// This should never fail, when it does, panic is thrown
func ShortHostname() (host string) {
	var err error
	if host, err = os.Hostname(); err != nil {
		panic(parl.Errorf("os.Hostname: '%w'", err))
	}
	if index := strings.Index(host, "."); index != -1 {
		host = host[:index]
	}
	return
}
