/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omap1

type traverser[K comparable, V any] struct {
	isBackward bool
	pair       *mappingNode[K, V]
}

func newTraverser[K comparable, V any](pair *mappingNode[K, V], back ...bool) (t *traverser[K, V]) {
	t = &traverser[K, V]{
		pair: pair,
	}
	if len(back) > 0 && back[0] {
		t.isBackward = true
	}
	return
}

func (t *traverser[K, V]) traverse(yield func(key K, value V) (keepGoing bool)) {
	var p = t.pair
	t.pair = nil
	for p != nil {
		if !yield(p.Key, p.Value) {
			return
		}
		if !t.isBackward {
			p = p.Next
		} else {
			p = p.Prev
		}
	}
}

const (
	backwards = true
)
