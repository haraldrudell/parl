/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package imap1

type ValueIterator[K comparable, V any] struct {
	m          IndexedMap[K, V]
	startIndex int
}

type KeyAndIndex[K any] struct {
	Key   K
	Index int
}

// Keys provides allocation-free iteration of [IndexedMap] values
func Values[K comparable, V any](m IndexedMap[K, V], startIndex ...int) (r ValueIterator[K, V]) {
	var index int
	if len(startIndex) > 0 {
		index = startIndex[0]
	} else {
		index = noStartIndex
	}
	r = ValueIterator[K, V]{m: m, startIndex: index}
	return
}

func (r ValueIterator[K, V]) R(yield func(value V, keyAndIndex KeyAndIndex[K]) (keepGoing bool)) {
	var m = r.m
	var i int
	if r.startIndex >= m.Length() {
		return
	} else if r.startIndex > 0 {
		i = r.startIndex
	}
	for i < m.Length() {
		var kai = KeyAndIndex[K]{Index: i}
		kai.Key, _ = m.GetKeyByIndex(i)
		var value, _ = /*hasValue*/ m.Get(kai.Key)
		if !yield(value, kai) {
			return
		}
		i++
	}
}

func (r ValueIterator[K, V]) Reverse(yield func(value V, keyAndIndex KeyAndIndex[K]) (keepGoing bool)) {
	var m = r.m
	var i int
	if r.startIndex >= m.Length() {
		return
	} else if r.startIndex > 0 {
		i = r.startIndex
	}
	for i >= 0 {
		var kai = KeyAndIndex[K]{Index: i}
		kai.Key, _ = m.GetKeyByIndex(i)
		var value, _ = /*hasValue*/ m.Get(kai.Key)
		if !yield(value, kai) {
			return
		}
		i--
	}
}

const (
	noStartIndex = -1
)
