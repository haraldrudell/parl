/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parlos provides simplified functions related to the os package
package pos

import (
	"errors"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// UserHomeDir obtains the absolute path to the process owning user’s
// home directory.
// This should never fail, when it does, panic is thrown
func UserHomeDir() (homeDir string) {

	// try getting home directory from account configuration
	homeDir = getProcessOwnerHomeDir()

	// if that fails, try shell environment
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir() // use $HOME environment variable
		if err != nil {
			panic(perrors.Errorf("os.UserHomeDir: '%w'", err))
		}
		if homeDir == "" {
			panic(perrors.New("failed to obtain home directory"))
		}
	}
	return
}

// HomeDir creates levels of directories in users’s home.
// if directories do not exist, they are created with permissions u=rwx.
// This should never fail, when it does, panic is thrown
func HomeDir(relPaths string) (dir string) {
	homeDir := UserHomeDir()
	dir = path.Join(homeDir, relPaths)
	if err := os.MkdirAll(dir, 0700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			panic(perrors.Errorf("os.MkdirAll: %w", err))
		}
	}
	return
}

// getProcessOwnerHomeDir retrives a user’s home directory
// based on account configuration.
// This is required for Linux system services that do not
// have an environment
// Best effort: errors are ignored
func getProcessOwnerHomeDir() (homeDir string) {

	// get process user ID
	userID := os.Geteuid()
	if userID == -1 { // on Windows, -1 is returned
		return // FAIL: user ID not found
	}

	// lookup the user ID
	userdata, err := user.LookupId(strconv.Itoa(userID))
	if err != nil {
		return // FAIL: user data not found
	}

	return userdata.HomeDir // path to the user's home directory
}

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
