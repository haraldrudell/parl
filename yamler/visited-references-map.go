/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package yamler

import (
	"reflect"
	"strings"

	"github.com/haraldrudell/parl/perrors"
	"gopkg.in/yaml.v3"
)

const (
	anyString = "interface\x20{}"
)

// structMapPath contains a struct-map[string]any pair and its path
//   - struct1 is traversed and the map2 is checked for fields existing meaning they are visited
type structMapPath struct {
	// y is the struct’s type
	struct1 *reflect.Value
	// - map is free-form unmarshaled yaml: any, map[string]any, []any
	map2 *reflect.Value
	// - path is dot-separated field names from the root struct
	//	- the root YamlData is empty slice
	//	- path then consists of lower-case field-names leadig to each sub-struct pointer
	path []string
}

// VisitedReferencesMap unmarshals yaml to an any object and then
// build a visited references map by comparring that object to its
// value pointer
func (u *Unmarshaler[T]) VisitedReferencesMap(yamlText []byte, yamlDictionaryKey string) (yamlVisistedReferences map[any]string, err error) {

	// unmarshal a freeform object: any, map[string]any, []any
	//	- by traversing values, we can determine which main.y fields
	//		were visited
	var freeFormYaml any
	if err = yaml.Unmarshal(yamlText, &freeFormYaml); perrors.IsPF(&err, "yaml.Unmarshal %w", err) {
		return // unmarshal error, should not happen
	}

	// obtain the options value
	var freeformValue any
	if topLevelDictionary, ok := freeFormYaml.(map[string]any); !ok {
		err = perrors.NewPF("yaml top-level object not dictionary")
		return
	} else if freeformValue = topLevelDictionary[yamlDictionaryKey]; freeformValue == nil {
		yamlVisistedReferences = make(map[any]string)
		return // options value not present: empty map return
	}
	yamlText = nil
	freeFormYaml = nil

	// now company u.y and freeFormValue to get referenced fields of y
	u.refs = make(map[any]string)
	if err = u.compareStructs(u.y, freeformValue); err != nil {
		return
	}

	yamlVisistedReferences = u.refs
	u.refs = nil

	return
}

// compareStructs compare a struct and map using reflection
//   - y is u.y pointer
//   - anyYaml is unmarshaled free-form yaml: any map[string]any []any
func (u *Unmarshaler[T]) compareStructs(y, anyYaml any) (err error) {

	// the structMapPath being processed
	var struct0 structMapPath
	if struct0.struct1, err = u.structp(y); perrors.Is(&err, "y: %w", err) {
		return
	}
	if struct0.map2, err = u.mapStringAny(anyYaml); perrors.Is(&err, "anyYaml: %w", err) {
		return
	}

	// list of struct-map-path values being processed
	var structs = []structMapPath{struct0}

	// process structs
	for len(structs) > 0 {
		struct0 = structs[0]
		structs = structs[1:]

		// iterate over fields
		var fieldCount = struct0.struct1.NumField()
		for i := 0; i < fieldCount; i++ {

			// field from struct
			var field = struct0.struct1.Field(i)
			// fieldName is lower-case as in yaml: “fieldone” not ”FieldOne”
			var fieldName = strings.ToLower(struct0.struct1.Type().Field(i).Name)
			if fieldName == "" {
				perrors.NewPF("")
			}
			var fieldKind = field.Kind()

			// field from map: mapValue is the any type
			var mapValue = struct0.map2.MapIndex(reflect.ValueOf(fieldName))
			if !mapValue.IsValid() {
				continue // freeFormYaml does not have the field
			}
			// runtime type for map value, ie. drop the interface {}
			var mapTypedValue = mapValue.Elem()

			// check if struct pointer
			if fieldKind == reflect.Pointer && mapTypedValue.CanInterface() {
				if structp, e1 := u.structp(field.Interface()); e1 == nil {
					if mp, e2 := u.mapStringAny(mapTypedValue.Interface()); e2 == nil {
						structs = append(structs, structMapPath{
							struct1: structp,
							map2:    mp,
							path:    append(append([]string{}, struct0.path...), fieldName), // clone
						})
						continue
					}
				}
			}

			// check that value types match
			if mapKind := mapTypedValue.Kind(); mapKind != fieldKind {
				err = perrors.ErrorfPF("field types different for path %q: struct: %q map: %q",
					strings.Join(append(append([]string{}, struct0.path...), fieldName), "."),
					fieldKind, mapKind,
				)
				return // value for field of different types return
			}

			// store referenced field pointer in map

			// make reference to field
			var fieldAddr = field.Addr()
			// make type any
			var anyPointerToField = fieldAddr.Interface()
			u.refs[anyPointerToField] = fieldName
		}
	}

	return
}

// mapStringAny ensures m is map[string]any using reflection or error
func (u *Unmarshaler[T]) mapStringAny(m any) (reflectMapStringAny *reflect.Value, err error) {

	// get non-nil reflect type of m
	var reflectTypeM = reflect.TypeOf(m)
	if reflectTypeM == nil {
		err = perrors.NewPF("m cannot be nil, must be map[string]any")
		return
	} else if reflectTypeM.Kind() != reflect.Map ||
		reflectTypeM.Key().Kind() != reflect.String ||
		reflectTypeM.Elem().String() != anyString {
		err = perrors.ErrorfPF("m must be map[string]any: %T", m)
		return
	}

	// get reflect value of m
	var reflectValueM = reflect.ValueOf(m)
	reflectMapStringAny = &reflectValueM

	return
}

// structp ensures v is pointer to struct using reflection or error
func (u *Unmarshaler[T]) structp(v any) (reflectStruct *reflect.Value, err error) {

	// get the non-nil reflect value of v
	var reflectValueV = reflect.ValueOf(v)
	if !reflectValueV.IsValid() {
		err = perrors.NewPF("v cannot be nil, must be *struct")
		return
	}

	// get the struct that supposedly v points to
	var structValue reflect.Value
	if reflectValueV.Kind() != reflect.Pointer {
		err = perrors.NewPF("v not pointer, must be *struct")
		return
	}
	structValue = reflectValueV.Elem()
	if structValue.Kind() != reflect.Struct {
		err = perrors.ErrorfPF("v points to non-struct value, must be *struct")
		return
	}

	reflectStruct = &structValue

	return
}
