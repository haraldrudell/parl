/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"reflect"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

type YamlValue struct {
	Name    string
	Pointer interface{}
}

//   - instancePointer is a reference to a YamlData value that must be struct
//   - fieldPointer is a list of references to values read from yaml
//     that matches the fields of YamlData
func NewYamlValue(instancePointer interface{}, fieldPointer interface{}) (yv *YamlValue) {

	// get the reflect value of the YamlData value struct
	//	- this is provided by instancePointer
	reflectValue := reflect.ValueOf(instancePointer)
	if !reflectValue.IsValid() {
		perrors.Errorf("NewYamlValue: instancePointer cannot be nil")
	}
	structValue := reflectValue.Elem() // retrieve what the pointer points to
	if structValue.Kind() != reflect.Struct {
		perrors.Errorf("NewYamlValue: instancePointer not pointer to struct instance")
	}

	// iterate over all fields of the value, ie. instancePointer
	numField := structValue.NumField()
	for i := 0; i < numField; i++ {
		fieldValue := structValue.Field(i)
		if !fieldValue.CanAddr() {
			perrors.Errorf("NewYamlValue: field#%d: canAddr false", i)
		}
		fieldAddr := fieldValue.Addr()
		if !fieldAddr.CanInterface() {
			perrors.Errorf("NewYamlValue: field#%d: canInterface false", i)
		}
		ifValue := fieldAddr.Interface()
		if ifValue == fieldPointer {
			return &YamlValue{
				Name:    strings.ToLower(structValue.Type().Field(i).Name),
				Pointer: fieldPointer}
		}
	}
	perrors.Errorf("NewYamlValue: fieldPointer not found")
	return
}
