/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"os"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

// ShortHostname gets hostname without domain part
// This should never fail, when it does, panic is thrown
func ShortHostname() (host string) {
	var err error
	if host, err = Hostname(); err != nil {
		panic(err)
	}

	return
}

// hostname without domain part
func Hostname() (host string, err error) {
	if host, err = os.Hostname(); perrors.IsPF(&err, "os.Hostname: %w", err) {
		return
	} else if index := strings.Index(host, "."); index != -1 {
		host = host[:index]
	}

	return
}
