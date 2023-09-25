/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"github.com/haraldrudell/parl/perrors"
)

const (
	DropPathsNotValues = true
)

// PathRegistry stores values that are accessible by index or by
// absolute path
type PathRegistry[V any] struct {
	paths  map[string]*V
	values []*pathTuple[V]
}

// pathTuple holds a value and the key for that value in the paths map
type pathTuple[V any] struct {
	abs   string
	value *V
}

// NewPathRegistry returns a regustry of values by absolute path
func NewPathRegistry[V any]() (registry *PathRegistry[V]) {
	return &PathRegistry[V]{paths: make(map[string]*V)}
}

// Add adds a value to the registry
func (r *PathRegistry[V]) Add(abs string, value *V) {
	r.paths[abs] = value
	r.values = append(r.values, &pathTuple[V]{
		abs:   abs,
		value: value,
	})
}

// HasAbs check whether an absolute path is stored in the registry
// as a key to a value
func (r *PathRegistry[V]) HasAbs(abs string) (hasAbs bool) {
	_, hasAbs = r.paths[abs]
	return
}

// ListLength returns the length of the value slice
//   - a value can still be nil for a discarded root
func (r *PathRegistry[V]) ListLength() (length int) {
	return len(r.values)
}

// MapLength returns the number of stored values
func (r *PathRegistry[V]) MapLength() (length int) {
	return len(r.paths)
}

// GetNext gets the next value and removes it from
// the registry
func (r *PathRegistry[V]) GetNext() (value *V) {

	for {

		// end of paths case
		if len(r.values) == 0 {
			return
		}

		// remove values from slice and map
		var tuple = r.values[0]
		r.values = r.values[1:]
		if tuple == nil {
			continue // empty cell
		}
		delete(r.paths, tuple.abs)

		value = tuple.value

		return
	}
}

// GetValue retrieves value by index
//   - if index is less than 0 or too large or for a removed
//     value, nil is returned
func (r *PathRegistry[V]) GetValue(index int) (value *V) {
	if index >= 0 && index < len(r.values) {
		if tp := r.values[index]; tp != nil {
			value = tp.value
		}
	}
	return
}

func (r *PathRegistry[V]) DeleteByIndex(index int) {
	if index < 0 || index >= len(r.values) {
		panic(perrors.ErrorfPF("delete index %d of 0…%d", index, len(r.values)-1))
	}
	var tuple = r.values[index]
	if tuple == nil {
		panic(perrors.ErrorfPF("delete of nil tuple at index %d", index))
	}
	delete(r.paths, tuple.abs)
	r.values[index] = nil

	// slice away nils at beginning and end
	var i0 = 0
	var i1 = len(r.values)
	for i0 < i1 && r.values[i0] == nil {
		i0++
	}
	for i1 > i0 && r.values[i1-1] == nil {
		i1--
	}
	if i0 != 0 || i1 != len(r.values) {
		r.values = r.values[i0:i1]
	}
	return
}

func (r *PathRegistry[V]) Drop(onlyPaths ...bool) {
	r.paths = nil
	if len(onlyPaths) == 0 || !onlyPaths[0] {
		r.values = nil
	}
}
