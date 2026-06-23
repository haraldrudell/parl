// © 2026–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License

package pslices

import (
	"fmt"
	"strconv"
	"strings"
)

// Format does Sprintf %v on slice without slow reflection
//   - inputSlice → “[1 2 3]”
func Format(inputSlice []byte) (numericStringRepresentation string) {

	// empty case
	if inputSlice == nil {
		numericStringRepresentation = "nil[]"
		return
	} else if len(inputSlice) == 0 {
		numericStringRepresentation = "[]"
		return
	}

	var b strings.Builder
	// Pre-allocate buffer space to eliminate intermediate resize allocations
	//	- each byte is 1–3 digits and one space: 4
	//	- 2 characters “[]”
	b.Grow(len(inputSlice)*4 + 2)
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(int(inputSlice[0])))
	for _, byteValue := range inputSlice[1:] {
		b.WriteByte('\x20')
		b.WriteString(strconv.Itoa(int(byteValue)))
	}
	b.WriteByte(']')
	numericStringRepresentation = b.String()

	return
}

func FormatStringer[Stringer fmt.Stringer](inputSlice []Stringer) (numericStringRepresentation string) {
	// empty case
	if inputSlice == nil {
		numericStringRepresentation = "nil[]"
		return
	} else if len(inputSlice) == 0 {
		numericStringRepresentation = "[]"
		return
	}

	var b strings.Builder
	// Pre-allocate buffer space to eliminate intermediate resize allocations
	//	- each byte is 1–3 digits and one space: 4
	//	- 2 characters “[]”
	b.WriteByte('[')
	b.WriteString(inputSlice[0].String())
	for _, stringer := range inputSlice[1:] {
		b.WriteByte('\x20')
		b.WriteString(stringer.String())
	}
	b.WriteByte(']')
	numericStringRepresentation = b.String()

	return
}
