/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parlos provides simplified functions related to the os package
package parlos

import (
	"errors"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/error116"
)

func UserHomeDir() (homeDir string) {
	homeDir = getProcessOwnerHomeDir()
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir() // use $HOME environment variable
		if err != nil {
			panic(error116.Errorf("os.UserHomeDir: '%w'", err))
		}
		if homeDir == "" {
			panic(error116.New("failed to obtain home directory"))
		}
	}
	return
}

// HomeDir creates levels of directories in users’s home
func HomeDir(relPaths string) (dir string) {
	homeDir := UserHomeDir()
	dir = path.Join(homeDir, relPaths)
	if err := os.MkdirAll(dir, 0700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			panic(error116.Errorf("os.MkdirAll: %w", err))
		}
	}
	return
}

func getProcessOwnerHomeDir() (homeDir string) {
	userID := os.Geteuid()
	if userID == -1 { // on Windows, -1 is returned
		return
	}
	userdata, err := user.LookupId(strconv.Itoa(userID))
	if err == nil {
		homeDir = userdata.HomeDir
	}
	return
}

// ShortHostname gets hostname without domain part
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
