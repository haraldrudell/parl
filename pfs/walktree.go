/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"
	"strings"
)

const (
	sSep = string(filepath.Separator)
)

// Tree consists of one or more indexed roots, each root a subhierarchy of the file system
type Tree struct {
	List  []Root
	Index map[string]int // key: evaled path, value: index in RootList
}

// NewTree allocates a Tree
func NewTree() *Tree {
	return &Tree{Index: map[string]int{}}
}

// AddRoot adds a new root
func (tree *Tree) AddRoot(path string) (symlinkTargets []string, err error) {
	root, err := NewRoot(path)
	if err != nil {
		return
	}

	// add new root node
	listIndex := len(tree.List)
	tree.Index[root.Evaled] = listIndex
	tree.List = append(tree.List, *root)
	root = &tree.List[listIndex]
	symlinkAggregator := func(path string) {
		symlinkTargets = append(symlinkTargets, path)
	}
	root.Build(symlinkAggregator)
	return
}

// ResolveSymlink patches the tree for a symlink
func (tree *Tree) ResolveSymlink(path string) (symlinkTargets []string, err error) {
	var abs string
	if abs, err = AbsEval(path); err != nil {
		return
	}
	exists, roots := tree.checkPath(abs)
	if exists {
		return // known root or subdirectory of known root
	}
	// TODO this discards work, processed roots should be used in traversing superior root
	for _, root := range roots {
		index := tree.Index[root.Evaled]
		delete(tree.Index, root.Evaled)
		for key, ix := range tree.Index {
			if ix < index {
				continue
			}
			tree.Index[key] = ix - 1
		}
		tree.List = append(tree.List[0:index], tree.List[index+1:]...)
	}
	return tree.AddRoot(path)
}

func (tree *Tree) checkPath(abs string) (exists bool, roots []*Root) {
	// is path an existing root or a subdirecty of an existing root?
	if _, exists = tree.Index[abs]; exists {
		return // abs is an existing root: exists true
	}
	for _, root2 := range tree.List {
		if strings.HasPrefix(abs+sSep, root2.Evaled+sSep) {
			exists = true
			return // abs is a subdirectory of an existing root: exists true
		}
		if strings.HasPrefix(root2.Evaled+sSep, abs+sSep) {
			roots = append(roots, &root2)
		}
	}
	return // abs is separate from all known roots: exists false, roots: sub-roots
}

// Walk traverses the built tree
func (tree *Tree) Walk(walkFunc filepath.WalkFunc) error {
	for _, root := range tree.List {
		if err := root.Walk(walkFunc); err != nil && err != filepath.SkipDir {
			return err
		}
	}
	return nil
}
