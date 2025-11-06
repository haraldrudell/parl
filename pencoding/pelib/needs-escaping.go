/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pelib

import (
	"encoding/xml"
	"unicode/utf8"
)

// NeedsEscaping escapes the minimum characters
//   - s is character data obtained from [xml.Decoder.Token]
//   - needs true: contains ampersand, less-than, tab, return or
//     out-of-range Unicode
//   - —
//   - — all escapes have been resolved
//   - — what encoding/xml unescapes:
//     ‘"’ ‘&#34;’ double-quote,
//     ‘'’ ‘&#39;’ apostrophe,
//     ‘&’ ‘&amp;’ ampersand,
//     ‘<’ ‘&lt;’ less-than,
//     ‘>’ ‘&gt;’ greater-than,
//     9 ‘&#x9;’ tab,
//     10 ‘&#xA;’ newline,
//     not encoded: 13 ‘&#xD;’ return,
//     some Unicode becomes replacement characters
func NeedsEscaping(s xml.CharData) (needs bool) {
	for i := 0; i < len(s); {

		// get next rune
		var r, width = utf8.DecodeRune(s[i:])
		i += width

		switch r {
		case '&', '<', '\t', '\r':
			needs = true
			return
		default:
			if !isInCharacterRange(r) || (r == 0xFFFD && width == 1) {
				needs = true
				return
			}
		}
	}

	return
}

// Decide whether the given rune is in the XML Character Range, per
// the Char production of https://www.xml.com/axml/testaxml.htm,
// Section 2.2 Characters.
func isInCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}
