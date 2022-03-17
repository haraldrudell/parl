/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

/*
filename: program-host.yaml program.yaml
Directories: ~/apps .. /etc
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/parlos"
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

type UnmarshalFunc func(in []byte, out interface{}) (err error) // yaml.Unmarshal
type UnmarshalThunk func(bytes []byte, unmarshal UnmarshalFunc, yamlKey string) (hasDate bool, err error)

// FindFile locates and read the yaml file
func FindFile(filename0, program string) (out string, bytes []byte) {
	if filename0 != "" {
		out = Abs(filename0)
		bytes, _ = readFile(out)
		return
	}
	dirs := []string{path.Join(parlos.UserHomeDir(), appsName), ParentDir(), Abs(etcName)}
	filenames := []string{fmt.Sprintf("%s-%s%s", program, parlos.ShortHostname(), yamlExt),
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
			panic(parl.Errorf("ioutil.ReadFile: '%w'", err))
		}
	} else {
		exists = true
	}
	return
}
