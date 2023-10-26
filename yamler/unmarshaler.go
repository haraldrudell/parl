/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"gopkg.in/yaml.v3"
)

// Unmarshaler is an object that can unmarshal yaml
// for unimported types
type Unmarshaler[T any] struct {
	y *T // the value pointer y and yaml object type T
	// a map of references to fields of y that appeared in yaml
	refs map[any]string
}

// NewUnmarshaler returns an object that can unmarshal yaml
// for unimported types
func NewUnmarshaler[T any](y *T) (unmarshaler GenericYaml) {
	return &Unmarshaler[T]{y: y}
}

// Unmarshal updates [yamlo.Unmarshaler]’s value pointer with
// data from yaml
//   - yamlText is utf8-encoded binary data read from the yaml file
//   - yamlDictionaryKey is the name of the top-level dictionary-key
//     typically “options”
//   - hasData indicates that unmarshal succeeded and yamlDictionaryKey
//     had value
func (u *Unmarshaler[T]) Unmarshal(yamlText []byte, yamlDictionaryKey string) (hasData bool, err error) {

	// top-level object in yaml is supposed to be a dictionary
	//	- therefore, wrap the options type T in map[string]
	//	- this makes T a value in a yaml dictionary
	//	- then check for the correct map key
	var yamlContentObject = map[string]*T{}

	// unmarshal yaml into the options object
	if err = yaml.Unmarshal(yamlText, &yamlContentObject); perrors.IsPF(&err, "yaml.Unmarshal", err) {
		return // unmarshal error return
	}

	// get the value of the T type read from yaml if it exists
	var yamlDataPointer = yamlContentObject[yamlDictionaryKey] // pick out the options dictionary value
	if hasData = yamlDataPointer != nil; !hasData {
		return // value for dictionary key not present return
	}

	// assign the unmarshaled yaml values to effective yaml values
	*u.y = *yamlDataPointer

	return // yaml loaded successfully return
}

// YDump returns field names and values for the yaml value struct
func (u *Unmarshaler[T]) YDump() (yamlVPrint string) {
	return parl.Sprintf("%+v", *u.y)
}
