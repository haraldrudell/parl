/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

/*
filename: program-host.yaml program.yaml
Directories: ~/apps .. /etc
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/pos"
)

const (
	emptyDir      = "."
	appsName      = "apps"
	etcName       = "/etc"
	yamlExt       = ".yaml"
	defaultTopKey = "options"
)

// GetTopLevelKey gets the top level key to use
func GetTopLevelKey(topKey string) (key string) {
	key = topKey
	if key == "" {
		key = defaultTopKey
	}
	return
}

// FindFile locates and read the yaml file
func FindFile(filename0, program string) (out string, bytes []byte) {
	if filename0 != "" {
		out = pfs.Abs(filename0)
		bytes, _ = readFile(out)
		return
	}
	dirs := []string{path.Join(pos.UserHomeDir(), appsName), pos.ParentDir(), pfs.Abs(etcName)}
	filenames := []string{fmt.Sprintf("%s-%s%s", program, pos.ShortHostname(), yamlExt),
		program + yamlExt}
	for _, dir := range dirs {
		for _, f := range filenames {
			filename := path.Join(dir, f)
			var exists bool
			if bytes, exists = readFile(filename); exists {
				out = filename
				return
			}
		}
	}
	return
}

func readFile(filename string) (bytes []byte, exists bool) {
	var err error
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		if !os.IsNotExist(err) {
			panic(perrors.Errorf("ioutil.ReadFile: '%w'", err))
		}
	} else {
		exists = true
	}
	return
}
