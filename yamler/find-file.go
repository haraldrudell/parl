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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
	"github.com/haraldrudell/parl/pos"
)

const (
	emptyDir = "."
	appsName = "apps"
	etcName  = "/etc"
	yamlExt  = ".yaml"
)

// FindFile locates and reads the yaml file
//   - if filename is empty, try dirs: [~/app, .., /etc] files: [app-host.yaml, app.yaml]
//   - readFilename is absolute, cleaned filename
func FindFile(filename, program string) (readFilename string, byts []byte, err error) {

	// try provided filename, must exist
	if filename != "" {
		if filename, err = filepath.Abs(filename); perrors.IsPF(&err, "filepath.Abs %w", err) {
			return // filename bad
		}
		if byts, _, err = readFile(filename); err == nil {
			readFilename = filename
		}
		return // filename set return
	}

	// [~/app, .., /etc]
	var dirs = []string{path.Join(pos.UserHomeDir(), appsName), pos.ParentDir(), pfs.Abs(etcName)}
	// [app-host.yaml, app.yaml]
	var filenames = []string{
		fmt.Sprintf("%s-%s%s", program, pos.ShortHostname(), yamlExt),
		program + yamlExt,
	}
	var doesNotExist bool
	for _, dir := range dirs {
		for _, f := range filenames {
			filename = path.Join(dir, f)
			if byts, doesNotExist, err = readFile(filename); err == nil {
				readFilename = filename
				return // successful read return
			} else if !doesNotExist {
				return // some error return
			}
			err = nil // ignore file not found error
		}
	}
	return // no file existed return
}

// readFile read the contents of the file fiename
//   - if err is nil, byts has content
//   - if err is non-nil and doesNotExist is true, the file does not exist
//   - if err is non-nil and doesNotExist is false, an error occurred
func readFile(filename string) (byts []byte, doesNotExist bool, err error) {
	if byts, err = os.ReadFile(filename); perrors.IsPF(&err, "os.ReadFile: %w", err) {
		doesNotExist = errors.Is(err, fs.ErrNotExist)
	}
	return
}
