/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlfs

import (
	"path/filepath"

	"github.com/haraldrudell/parl"
)

// FSEntry represnt a branch in a file system hierarchy
// FSEntry is implemented by Entry (file) or Directory or Root (walk entry point)
type FSEntry interface {
	SafeName() string
	IsDir() bool
	Walk(path string, walkFunc filepath.WalkFunc) error
}

// Root is a file system hierarchy
type Root struct {
	Path    string // path as provided
	Evaled  string // path after filepath.EvalSymlinks
	Err     error
	FSEntry // Directory or Entry
}

// NewRoot allocates a root
func NewRoot(path string) (rt *Root, err error) {
	root := Root{Path: path}
	if root.Evaled, err = AbsEval(path); err != nil {
		return
	}
	rt = &root
	return
}

// Build scans the file system for this root
func (rt *Root) Build(sym func(string)) {
	if rt.FSEntry != nil {
		panic(parl.New("root.Build invoked more than once"))
	}
	rt.FSEntry = NewEntry(rt.Path, "", sym)
}

// Walk traverses the root
func (rt *Root) Walk(walkFunc filepath.WalkFunc) error {
	if rt.FSEntry == nil {
		panic(parl.New("root.Walk without root.Build"))
	}

	// the root node is special, because the entire path is already known
	return rt.FSEntry.Walk(rt.Path, walkFunc)
}
