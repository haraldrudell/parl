/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pbytes

import "bytes"

var pbNewlines = [][]byte{
	[]byte("\r\n"),
	[]byte("\n"),
	[]byte("\r"),
}

func TrimNewline(in []byte) (out []byte) {
	for _, nl := range pbNewlines {
		if out = bytes.TrimSuffix(in, nl); len(in) != len(out) {
			return
		}
	}
	out = in
	return
}
