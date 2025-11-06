/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pelib

import (
	"bytes"
	"encoding/xml"
	"io"

	"github.com/haraldrudell/parl/perrors"
)

type Tokenizer struct {
	// dec must be heap-allocated
	//	- must use [xml.NewDecoder] to initialize
	//		package-private fields: 1 allocation
	dec *xml.Decoder
}

// MakeTokenizer provides the Token method
//   - reader: heao-allocated byte-reader
func MakeTokenizer(reader io.Reader) (tokenizer Tokenizer) {
	return Tokenizer{
		dec: xml.NewDecoder(reader),
	}
}

// Token returns an xml token stream removing inner XML white-space
func (t *Tokenizer) Token() (token xml.Token, tokenType TokenType, err error) {

	for {

		// read next XML token: element, character data, comment,
		// processing instruction or namespace declaration
		if token, err = t.dec.Token(); err != nil {
			if err == io.EOF {
				tokenType = IsEOF
				err = nil
			} else {
				err = perrors.ErrorfPF("parse xml %w", err)
			}
			// decoding error
			return
		}

		// check if inner xml characters
		var charData, ok = token.(xml.CharData)
		if !ok {
			// token is not inner xml

			if _, isStartElement := token.(xml.StartElement); isStartElement {
				tokenType = IsOpenTag
			} else if _, isEndElement := token.(xml.EndElement); isEndElement {
				tokenType = IsCloseTag
			} else if _, isProcInst := token.(xml.ProcInst); isProcInst {
				tokenType = IsProcInst
			}
			return
		}

		// check for white-space-only inner xml characters
		//	- remove leading and trailing Unicode white-space
		var trimmed = bytes.TrimSpace(charData)
		if len(trimmed) > 0 {
			// not white-space-only character data
			return
		}
	}
}
