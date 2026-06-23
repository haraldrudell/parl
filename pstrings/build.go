// © 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License

package pstrings

import "strings"

func Build(stringList ...string) (spacedString string) {

	// empty case
	if len(stringList) == 0 {
		return
	}

	var b strings.Builder
	// pre-allocate
	var size = len(stringList) - 1
	for _, s := range stringList {
		size += len(s)
	}
	b.Grow(size)

	for i, s := range stringList {
		if i > 0 {
			b.WriteByte('\x20')
		}
		b.WriteString(s)
	}
	spacedString = b.String()

	return
}
