/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

// Registry stores values that are accessible by index or by
// absolute path
type Registry[V any] struct {
	// path provides O(1) access to values
	//	- mappings are deleted by DeleteByIndex
	paths  map[string]*V
	values []*pathTuple[V]
}

// NewRegistry returns a registry of values by absolute path
func NewRegistry[V any]() (registry *Registry[V]) { return &Registry[V]{paths: make(map[string]*V)} }

// Add adds a value to the registry
func (r *Registry[V]) Add(abs string, value *V) {
	r.paths[abs] = value
	r.values = append(r.values, &pathTuple[V]{
		abs:   abs,
		value: value,
	})
}

// HasAbs check whether an absolute path is stored in the registry
// as a key to a value
func (r *Registry[V]) HasAbs(abs string) (hasAbs bool) {
	_, hasAbs = r.paths[abs]
	return
}

// ListLength returns the length of the value slice
//   - a value can still be nil for a discarded root
func (r *Registry[V]) ListLength() (length int) { return len(r.values) }

// GetValue retrieves value by index
//   - if index is less than 0 or too large or for a removed
//     value, nil is returned
func (r *Registry[V]) GetValue(index int) (value *V) {
	if index >= 0 && index < len(r.values) {
		if tp := r.values[index]; tp != nil {
			value = tp.value
		}
	}
	return
}

func (r *Registry[V]) ObsoleteIndex(index int) {
	if index < 0 && index >= len(r.values) {
		return
	}
	var tuple = r.values[index]
	tuple.value = nil
	delete(r.paths, tuple.abs)
}
