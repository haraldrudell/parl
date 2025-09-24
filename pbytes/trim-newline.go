/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pbytes

import "bytes"

// newlinc-character variants
var pbNewlines = [][]byte{
	[]byte("\r\n"), // Windows
	[]byte("\n"),   // Unix-like and macOS
	[]byte("\r"),   // legacy macOS
}

// TrimNewline trims Unix-like obsolete macOS and Windows newlines
func TrimNewline(in []byte) (out []byte) {
	for _, nl := range pbNewlines {
		if out = bytes.TrimSuffix(in, nl); len(in) != len(out) {
			return
		}
	}
	out = in

	return
}
