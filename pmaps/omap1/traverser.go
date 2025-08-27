/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omap1

// traverser iterates over the doubly-linked list of [OrderedMap]
type traverser[K comparable, V any] struct {
	isBackward bool
	node       *mappingNode[K, V]
}

// newTraverser returns an iterator over the doubly-linked list of [OrderedMap]
func newTraverser[K comparable, V any](node *mappingNode[K, V], back ...bool) (t *traverser[K, V]) {
	t = &traverser[K, V]{
		node: node,
	}
	if len(back) > 0 && back[0] {
		t.isBackward = true
	}
	return
}

func (t *traverser[K, V]) traverse(yield func(key K, value V) (keepGoing bool)) {
	var mapping = t.node
	t.node = nil
	for mapping != nil {
		if !yield(mapping.Key, mapping.Value) {
			return
		}
		if !t.isBackward {
			mapping = mapping.Next
		} else {
			mapping = mapping.Prev
		}
	}
}

const (
	backwards = true
)
