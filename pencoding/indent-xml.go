/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pencoding

import (
	"encoding/xml"

	"github.com/haraldrudell/parl/pencoding/pelib"
)

// IndentXml puts elements on new line and indents using two spaces
//   - xmlDocument: xml document, may be nil or empty
//   - indentedXml: xmlDocument indented with tags on separate lines
//   - — if no elements: nil
//   - — newline after leading procinst
//   - — one xml element per line, child-indentation two spaces
//   - — uses self-closing elements
//   - — unnecessary character-data escapes may be resolved
//   - — character data with ampersand, less-than, tab, return
//     or out-of-range Unicode is escaped by encoding/xml
//   - err: marshal/unmarshal errors
func IndentXml(xmlDocument []byte) (indentedXml []byte, err error) {

	// xmler decodes input xml into tokens
	// removing white-space-only inner xml character data
	var xmler = pelib.MakeXmler(xmlDocument)
	// output should be indented
	xmler.SetIndent()

	var lastWasOpenTag bool
	var isFirst = true
	for {

		// get next token
		var token xml.Token
		var tokenType pelib.TokenType
		if token, tokenType, err = xmler.Token(); err != nil {
			return
		} else if tokenType == pelib.IsEOF {
			break
		}

		// p(token, tokenType)

		// add newline after any leading procinst
		if isFirst {
			isFirst = false
			if tokenType == pelib.IsProcInst {
				if err = xmler.WriteProcInstZero(token); err != nil {
					return
				}
				// skip the written procinst
				continue
			}
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
	indentedXml, err = xmler.End()

	return
}

// func p(token xml.Token, tokenType TokenType) {

// 	var name xml.Name
// 	var xmlns string

// 	if se, isStart := token.(xml.StartElement); isStart {
// 		name = se.Name
// 		xmlns = ns(se.Attr)
// 	} else if ee, isEnd := token.(xml.EndElement); isEnd {
// 		name = ee.Name
// 	} else {
// 		return
// 	}

// 	var nameS string
// 	if name.Space != "" {
// 		nameS = fmt.Sprintf("%s:%s", name.Space, name.Local)
// 	} else {
// 		nameS = name.Local
// 	}
// 	if xmlns != "" {
// 		xmlns = "\x20" + xmlns
// 	}

// 	parl.D("Token: tokenType: %s ns:tag: %s%s", tokenType, nameS, xmlns)
// }

// func ns(attrs []xml.Attr) (xmlns string) {
// 	for attr := range piter.SlicePointers(attrs).R {
// 		if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
// 			xmlns = fmt.Sprintf("xmlns=%s", attr.Value)
// 			return
// 		}
// 	}

// 	return
// }
