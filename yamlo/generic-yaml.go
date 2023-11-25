/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamlo

// GenericYaml is a wrapper for [yamlo.Unmarshaler] that allows the
// yamlo package to unmarshal yaml to an unimported type
type GenericYaml interface {
	// Unmarshal updates [yamlo.Unmarshaler]’s value pointer with
	// data from yaml
	//	- yamlText is utf8-encoded binary data read from the yaml file
	//	- yamlDictionaryKey is the name of the top-level dictionary-key
	//		typically “options”
	//	- hasData indicates that unmarshal succeeded and yamlDictionaryKey
	//		was present
	Unmarshal(yamlText []byte, yamlDictionaryKey string) (hasData bool, err error)
	// VisitedReferencesMap returns a map of
	// key: any-typed pointers to fields of u.y,
	// value: lower-case field names
	// - unmarshals yaml again to an any object and then
	// builds the visited references map by comparing the unmarshaled object to the
	// u.y struct-pointer
	VisitedReferencesMap(yamlText []byte, yamlDictionaryKey string) (yamlVisistedReferences map[any]string, err error)
	// YDump returns field names and values for the yaml value struct
	YDump() (yamlVPrint string)
}
