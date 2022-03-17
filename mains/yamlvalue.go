/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"reflect"
	"strings"

	"github.com/haraldrudell/parl"
)

type YamlValue struct {
	Name    string
	Pointer interface{}
}

func NewYamlValue(instancePointer interface{}, fieldPointer interface{}) (yv *YamlValue) {
	reflectValue := reflect.ValueOf(instancePointer)
	if !reflectValue.IsValid() {
		parl.Errorf("NewYamlValue: instancePointer cannot be nil")
	}
	structValue := reflectValue.Elem() // retrieve what the pointer points to
	if structValue.Kind() != reflect.Struct {
		parl.Errorf("NewYamlValue: instancePointer not pointer to struct instance")
	}
	numField := structValue.NumField()
	for i := 0; i < numField; i++ {
		fieldValue := structValue.Field(i)
		if !fieldValue.CanAddr() {
			parl.Errorf("NewYamlValue: field#%d: canAddr false", i)
		}
		fieldAddr := fieldValue.Addr()
		if !fieldAddr.CanInterface() {
			parl.Errorf("NewYamlValue: field#%d: canInterface false", i)
		}
		ifValue := fieldAddr.Interface()
		if ifValue == fieldPointer {
			return &YamlValue{
				Name:    strings.ToLower(structValue.Type().Field(i).Name),
				Pointer: fieldPointer}
		}
	}
	parl.Errorf("NewYamlValue: fieldPointer not found")
	return
}
