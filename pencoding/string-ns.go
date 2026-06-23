/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pencoding

import "encoding/xml"

const (
	// AttrXmlns is the attribute name for default namespace
	AttrXmlns = "xmlns"
	// AttrXmlnsFormat is [fmt.Sprintf] template for defining namespace prefixes
	AttrXmlnsFormat = "xmlns:%s"
)

// StringerNs is like [fmt.Stringer] but provides
// default namespace so it can be abbreviated
type StringerNs interface {
	// StringNs receives the current default namespace
	// so it can be removed
	//	- defaultNamespace: “DAV:” “urn:ietf:params:xml:ns:carddav”
	//	- any namespace prefix like “C:addressbook-home-set” is on parsing
	//		using [xml.Unmarshal] resolved so that Space Local are fully qualified
	StringNs(defaultNamespace string) (s string)
}

// provide xml to GoDoc
var _ = xml.Unmarshal
