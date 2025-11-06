/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pencoding

import (
	"encoding/xml"

	"github.com/haraldrudell/parl/pencoding/pelib"
)

// MinifyXml removes whitespace from XML
//   - xmlDocument: xml document, may be nil or empty
//   - noWhiteSpaceXml: xmlDocument with white-space removed
//   - — if no elements: nil
//   - — uses self-closing elements
//   - — unnecessary character-data escapes may be resolved
//   - — character data with ampersand, less-than, tab, return
//     or out-of-range Unicode is escaped by encoding/xml
//   - err: marshal/unmarshal errors
func MinifyXml(xmlDocument []byte) (noWhiteSpaceXml []byte, err error) {

	// encoding/xml cannot parse arbitrary xml
	//	- instead decode into stream of xml tokens

	// encoding/xml expands input self-cloding elements into open-close elements
	//	- therefore make empty elements self-closing

	// xmler decodes input xml into tokens
	// removing white-space-only inner xml character data
	var xmler = pelib.MakeXmler(xmlDocument)

	var lastWasOpenTag bool
	for {

		// get next token
		var token xml.Token
		var tokenType pelib.TokenType
		if token, tokenType, err = xmler.Token(); err != nil {
			return
		} else if tokenType == pelib.IsEOF {
			break
		}

		// [xml.Encoder.EncodeToken] does not have default namespace detection
		// [xml.Encoder.Encode] does
		//	- because tokens come from decoder, EncodeToken must be used
		//	- therefore namespaces must be manually managed

		// build map of namespaces
		xmler.ProcessNamespace(token, tokenType)

		// check for self-closing element
		if lastWasOpenTag && tokenType == pelib.IsCloseTag {
			if err = xmler.MakeSelfClosing(token, tokenType); err != nil {
				return
			}
			// closing tag complete, get next token
			lastWasOpenTag = false
			continue
		}

		// write token
		if err = xmler.WriteToken(token, tokenType); err != nil {
			return
		}

		// update lastWasOpenTag
		lastWasOpenTag = tokenType == pelib.IsOpenTag
		if lastWasOpenTag {
			// get the output byte position
			if err = xmler.GetPosition(); err != nil {
				return
			}
		}
	}
	// xmlDocument read to end

	// flush xml to output buffer
	noWhiteSpaceXml, err = xmler.End()

	return
}
