/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

import (
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/yamlo"
	"golang.org/x/exp/maps"
)

func TestNewUnmarshaler(t *testing.T) {
	type Struct2 struct {
		FieldThree int
	}
	type YamlData struct {
		FieldOne, FieldTwo int
		FieldFour          []string
		P                  *Struct2
		// structs must be pointer P2                 Struct2
	}
	var y = YamlData{P: &Struct2{}}
	var yamlDictionaryKey = "options"
	var yamlText = []byte(`
options:
  fieldtwo: 3
  p:
    fieldthree: 4
  fieldfour:
  - someValue
`)
	var expMap = map[any]string{
		&y.FieldTwo:     "fieldtwo",
		&y.FieldFour:    "fieldfour",
		&y.P.FieldThree: "fieldthree",
	}

	var unmarshaler yamlo.GenericYaml
	var yamlVisistedReferences map[any]string
	var err error

	unmarshaler = NewUnmarshaler(&y)
	yamlVisistedReferences, err = unmarshaler.VisitedReferencesMap(yamlText, yamlDictionaryKey)
	if err != nil {
		t.Errorf("VisitedReferencesMap err: %s", perrors.Short(err))
	}
	if !maps.Equal(yamlVisistedReferences, expMap) {
		t.Errorf("map:\n%v exp:\n%v",
			yamlVisistedReferences, expMap,
		)
	}
}
