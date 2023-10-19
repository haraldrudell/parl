/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

// GenericYaml is a wrapper for [yamlo.Unmarshaler] that allows the
// yamler package to unmarshal yaml to an unimported type
type GenericYaml interface {
	// Unmarshal updates [yamlo.Unmarshaler]’s value pointer with
	// data from yaml
	//	- yamlText is utf8-encoded binary data read from the yaml file
	//	- yamlDictionaryKey is the name of the top-level dictionary-key
	//		typically “options”
	//	- hasData indicates that unmarshal succeeded and yamlDictionaryKey
	//		was present
	Unmarshal(yamlText []byte, yamlDictionaryKey string) (hasData bool, err error)
	// VisitedReferencesMap unmarshals yaml to an any object and then
	// build a visited references map by comparring that object to its
	// value pointer
	VisitedReferencesMap(yamlText []byte, yamlDictionaryKey string) (yamlVisistedReferences map[any]string, err error)
	YDump() (yamlVPrint string)
}
