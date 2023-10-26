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
	// the reflect.Type.String value for the any type
	anyString = "interface\x20{}"
)

// structMapPath is a work item comparing a struct with an unmarshaled
// free-form object consisting of values any, map[string] and and []any
//   - struct1 is traversed and the
//   - map2 is checked for the existence of struct1 fields
//   - meaning those fields were visited
//   - path is a field-name list for nested structs
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

// VisitedReferencesMap returns a map of
// key: any-typed pointers to fields of u.y,
// value: lower-case field names
// - unmarshals yaml again to an any object and then
// builds the visited references map by comparing the unmarshaled object to the
// u.y struct-pointer
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
		return // bad yaml return
	} else if freeformValue = topLevelDictionary[yamlDictionaryKey]; freeformValue == nil {
		yamlVisistedReferences = make(map[any]string)
		return // options value not present: empty map return
	}
	yamlText = nil
	freeFormYaml = nil

	// now compare u.y and freeFormValue to get referenced fields of y
	u.refs = make(map[any]string)
	if err = u.compareStructs(u.y, freeformValue); err != nil {
		return // nil or not struct* error during struct compare
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
		return // y nil or not struct*
	}
	if struct0.map2, err = u.mapStringAny(anyYaml); perrors.Is(&err, "anyYaml: %w", err) {
		return // anyYaml nil or not map[string]any
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
				err = perrors.ErrorfPF("field#d: field name empty", i)
				return // empty field name return
			}
			var fieldKind = field.Kind()

			// field from map: map[string]any returns the any type
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
						continue // structPair stired for processing, check next field
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
			// make type any (exits reflect domain)
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
		return // nil return
	} else if reflectTypeM.Kind() != reflect.Map ||
		reflectTypeM.Key().Kind() != reflect.String ||
		reflectTypeM.Elem().String() != anyString {
		err = perrors.ErrorfPF("m must be map[string]any: %T", m)
		return // not map[string]any return
	}

	// get reflect value of m
	var reflectValueM = reflect.ValueOf(m)
	reflectMapStringAny = &reflectValueM

	return // good return
}

// structp ensures v is pointer to struct using reflection, if not: error
//   - error on v nil or not *struct
func (u *Unmarshaler[T]) structp(v any) (reflectStruct *reflect.Value, err error) {

	// get the non-nil reflect value of v
	var reflectValueV = reflect.ValueOf(v)
	if !reflectValueV.IsValid() {
		err = perrors.NewPF("v cannot be nil, must be *struct")
		return // nil return
	}

	// get the struct that supposedly v points to
	var structValue reflect.Value
	if reflectValueV.Kind() != reflect.Pointer {
		err = perrors.NewPF("v not pointer, must be *struct")
		return // not pointer return
	}
	structValue = reflectValueV.Elem()
	if structValue.Kind() != reflect.Struct {
		err = perrors.ErrorfPF("v points to non-struct value, must be *struct")
		return // not struct* return
	}

	reflectStruct = &structValue

	return //  good return
}
