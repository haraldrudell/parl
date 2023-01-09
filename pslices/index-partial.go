/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"bytes"

	"golang.org/x/exp/slices"
)

// IndexPartial returns scan-result in a byte-slice for a key byte-sequence, whether key was found
// or more bytes needs to be read to determine if key is present.
//   - index is the first index to search at and it is checked against the length of byts
//   - return value:
//   - — notKnown: true, newIndex >= index: byts ended during a partial key match, newIndex is first byte of match
//   - — notKnown: false, newIndex >= index: entire key was found at newIndex
//   - — newIndex -1: no key sequence was found in byts
//   - empty or nil key is found immediately, but only if index is less than length of byts
//   - if byts is nil or empty, nothing is found
//   - panic-free
func IndexPartial(byts []byte, index int, key []byte) (newIndex int, notKnown bool) {
	if index < 0 {
		index = 0
	}
	if index < len(byts) {

		// check for entire key in byts
		// if key is empty, bytes.Index will find it
		if index+len(key) <= len(byts) {
			if newIndex = bytes.Index(byts[index:], key); newIndex != -1 {
				newIndex += index // make index based on byts
				return            // entire key found return: newIndex pos, notKnown: false
			}
		}

		// if key length is 1, key was not found
		if len(key) == 1 {
			return // 1-byte-key not found: newIndex -1, notKnown: false
		}

		// check for partial key at end of byts
		firstByte := key[:1]
		restOfKey := key[1:]
		if index+len(restOfKey) < len(byts) {
			index = len(byts) - len(restOfKey) // byts is long, search the last key-length - 1 bytes
		}
		for index < len(byts) {
			if newIndex = bytes.Index(byts[index:], firstByte); newIndex == -1 {
				return // not even partial key found return: newIndex: -1, notKnown: false
			}
			newIndex += index    // make index in byts
			index = newIndex + 1 // next byte to search
			if length := len(byts) - index; length > 0 {
				if slices.Compare(byts[index:index+length], restOfKey[:length]) != 0 {
					continue // restOfKey does not match, continue searching
				}
			}
			notKnown = true
			return // partial key found at end: newIndex: pos, notKnown: true
		}
	}
	newIndex = -1
	return // key not found return: newIndex: -1, notKnown: false
}
