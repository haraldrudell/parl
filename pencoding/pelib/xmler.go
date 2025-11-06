/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pelib

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"maps"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/piter"
)

// Xmler provides tokenizer, encoder, output buffer and glue-methods
type Xmler struct {
	// Token()
	//	 - allocates the encoder
	Tokenizer
	//
	//	- must be heap-allocated to provide to NewEncoder
	output *bytes.Buffer
	//	- must be heap-allocated to initialize package-private fields
	encoder *xml.Encoder
	// level is the numer of nested elements
	level int
	// nsList contains current namespace aliasing
	nsList []namespacer
	// byteCountAfterOpenElement remembers the
	// output location of the last opening tag
	byteCountAfterOpenElement int
}

// namespacer contains namespace mappings defined by
// opening tags
type namespacer struct {
	// the level at which namespaces were added
	//	- 0 is root element
	level int
	// key: resolved namespace, like “DAV:” “http://www.example.com/books”
	// value: namespace alias “C” or blank for default namepsace
	nsMap map[string]string
}

// MakeXmler returns a combined xml tokenizer-encoder
func MakeXmler(xmlDocument []byte) (x Xmler) {
	var output = bytes.NewBuffer(defaultBuffer)
	return Xmler{
		Tokenizer: MakeTokenizer(bytes.NewReader(xmlDocument)),
		output:    output,
		encoder:   xml.NewEncoder(output),
	}
}

// SetIndent enables newlines and indent in the output
func (x *Xmler) SetIndent() { x.encoder.Indent(ePrefix, eIndent) }

// WriteProcInstZero ensures a newline after initial process instruction
//   - invoked when first token is process instruction for
//     xml indentation
func (x *Xmler) WriteProcInstZero(token xml.Token) (err error) {

	if err = x.encoder.EncodeToken(token); perrors.IsPF(&err, "EncodeToken ProcInst %w", err) {
		return
	} else if err = x.encoder.Flush(); perrors.IsPF(&err, "Flush ProcInst %w", err) {
		return
	} else if _ /*n*/, err = x.output.Write(newlineByte); perrors.IsPF(&err, "WriteB %w", err) {
		return
	}

	return
}

// ProcessNamespace updates namespace aliasing
// when token is open or closing tag
func (x *Xmler) ProcessNamespace(token xml.Token, tokenType TokenType) {
	switch tokenType {
	case IsOpenTag:

		// it’s opening tag: may declare namespaces
		var se = token.(xml.StartElement)
		var nsMap map[string]string
		for attr := range piter.SlicePointers(se.Attr).R {
			// <… xmlns="DAV:"> → Space ‘’ Local “xmlns” Value “DAV:”
			// <… xmlns:C="x"> → Space ‘xmlns’ Local “C” Value “x”
			var isDefault = attr.Name.Space == "" && attr.Name.Local == xmlnsS
			var isMapped = attr.Name.Space == xmlnsS
			var nsAlias string
			if isMapped {
				// custom alias has specific alias name, “C” above
				nsAlias = attr.Name.Local
			} else if !isDefault {
				// not a namspace attribute
				continue
			}
			// found a namespace to be saved: nsAlias

			// ensure a map exists
			if nsMap == nil {
				if len(x.nsList) > 0 {
					nsMap = maps.Clone(x.nsList[len(x.nsList)-1].nsMap)
				} else {
					nsMap = make(map[string]string)
				}
				x.nsList = append(x.nsList, namespacer{
					level: x.level,
					nsMap: nsMap,
				})
			}

			// save namespace
			nsMap[attr.Value] = nsAlias
		}

		x.level++
		return
	case IsCloseTag:

		// closing tag discards alias maps
		//	- the alias map for the opening tag must be present
		//		during closing tag operation
		//	- thus means map zero will never be dropped
		x.level--

		// drop obsolete maps
		for len(x.nsList) > 0 {
			var index = len(x.nsList) - 1
			var lev = x.nsList[index].level
			// if x.level and lev is same, keep the map
			//	- if x.level less than lev, discard map
			if x.level >= lev {
				// this entry is still valid
				break
			}
			// drop the last map
			x.nsList[index].nsMap = nil
			x.nsList = x.nsList[:index]
		}
	}
}

// MakeSelfClosing makes the last open-tag self-closing
func (x *Xmler) MakeSelfClosing(token xml.Token, tokenType TokenType) (err error) {

	token = x.resolveNamespace(token, tokenType)

	// write closing token
	//	- encode/xml tracks and matches start-end, so this must be done
	//	- the closing tag will be discarded
	if err = x.encoder.EncodeToken(token); perrors.IsPF(&err, "EncodeToken %w", err) {
		return
		// flush close element to output
	} else if err = x.encoder.Flush(); perrors.IsPF(&err, "Flush %w", err) {
		return
	}
	// modify opening tag to be self-closing
	x.output.Truncate(x.byteCountAfterOpenElement - 1)
	if _ /*n*/, err = x.output.Write(slashGreaterThan); perrors.IsPF(&err, "Write %w", err) {
		return
	}

	return
}

// WriteToken writes a token to encoder or output buffer
func (x *Xmler) WriteToken(token xml.Token, tokenType TokenType) (err error) {

	// [xml.Encoder.EncodeToken] unnecessarily escapes
	//	- double-quote, apostrophe, greater-than and more
	//	- must be quoted: less-than, ampersand
	//	- [xml.CharData] is typed byte-slice
	//	- puropse is to avoid escaping of double-quote
	if cData, isCharacterData := token.(xml.CharData); isCharacterData {
		if !NeedsEscaping(cData) {

			// flush encoder to enable write to output
			if err = x.encoder.Flush(); perrors.IsPF(&err, "FlushB %w", err) {
				return
			} else if _, err = x.output.Write(cData); perrors.IsPF(&err, "WriteB %w", err) {
				return
			}
			// character data was written directly to output: complete
			return
		}
	}
	// should write using EncodeToken

	token = x.resolveNamespace(token, tokenType)

	if err = x.encoder.EncodeToken(token); perrors.IsPF(&err, "EncodeToken %w", err) {
		return
	}

	return
}

// GetPosition remembers the location of an opening tag
// so that it can later be made self-closing
func (x *Xmler) GetPosition() (err error) {
	if err = x.encoder.Flush(); perrors.IsPF(&err, "Flush %w", err) {
		return
	}
	x.byteCountAfterOpenElement = x.output.Len()

	return
}

// End closes the encoder and returns resulting xml document
func (x *Xmler) End() (xmlDocument []byte, err error) {

	// flush xml to output buffer
	if err = x.encoder.Close(); perrors.IsPF(&err, "xml Close %w", err) {
		return
	}
	xmlDocument = x.output.Bytes()

	return
}

// resolveNamespace performs namespace aliasing on an
// open or closing tag
//   - also patches attribute namespacing for opening tags
//   - because token interface returns value not pointer,
//     on change, token needs a new value
func (x *Xmler) resolveNamespace(token xml.Token, tokenType TokenType) (token2 xml.Token) {
	token2 = token

	// no namespace aliases available
	if len(x.nsList) == 0 {
		return
	}

	// resolve namespace
	//	- to get the namespace value,
	//		token needs to be type asserted
	var se xml.StartElement
	var ee xml.EndElement
	var space string
	switch tokenType {
	case IsOpenTag:
		se = token.(xml.StartElement)
		space = se.Name.Space
	case IsCloseTag:
		ee = token.(xml.EndElement)
		space = ee.Name.Space
	}

	// no namespace to alias in token to write
	if space == "" {
		return
	}

	// see if namespace has alias
	var alias string
	var wasFound bool
	var nsMap = x.nsList[len(x.nsList)-1].nsMap
	alias, wasFound = nsMap[space]

	// no change
	if !wasFound || alias == space {
		return
	}

	// open and close tags must have the same Space and tag
	// or encoding/xml returns ‘does not match’ error
	switch tokenType {
	case IsOpenTag:
		// the result should not have namespace
		//	- this would cause [xml.Encoder.EncodeToken] to insert xmlns attributes
		se.Name.Space = ""
		if alias != "" {
			// prepend namespace to tag name
			se.Name.Local = fmt.Sprintf("%s:%s", alias, se.Name.Local)
		}
		// custom attribute processing
		for attr := range piter.SlicePointers(se.Attr).R {
			if attr.Name.Space == xmlnsS {
				attr.Name.Local = fmt.Sprintf("%s:%s", xmlnsS, attr.Name.Local)
				attr.Name.Space = ""
			}
		}
		token2 = se
	case IsCloseTag:
		// the result should not have namespace
		//	- this would cause [xml.Encoder.EncodeToken] to insert xmlns attributes
		ee.Name.Space = ""
		if alias != "" {
			// prepend namespace to tag name
			ee.Name.Local = fmt.Sprintf("%s:%s", alias, ee.Name.Local)
		}
		token2 = ee
	}

	return
}

const (
	ePrefix = ""
	eIndent = "\x20\x20"
	xmlnsS  = "xmlns"
)

var (
	defaultBuffer    []byte
	newlineByte      = []byte{10}
	slashGreaterThan = []byte("/>")
)
