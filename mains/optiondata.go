/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

import (
	"flag"
	"fmt"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	silentOption = "silent"
	SilentString = "-" + silentOption
)

// OptionData contain options data for the flag package
type OptionData struct {
	P     interface{}
	Name  string
	Value interface{}
	Usage string
	Y     *YamlValue // pointer to data value from yaml
}

// AddOption executes flag.BoolVar and such on the options map
func (om *OptionData) AddOption() {
	switch p := om.P.(type) {
	case *bool:
		if value, ok := om.Value.(bool); !ok {
			panic(fmt.Errorf("option '%s' Boolean default value not boolean: %T", om.Name, om.Value))
		} else {
			flag.BoolVar(p, om.Name, value, om.Usage)
		}
	case *time.Duration:
		if value, ok := om.Value.(time.Duration); !ok {
			panic(fmt.Errorf("option '%s' Duration default value not duration: %T", om.Name, om.Value))
		} else {
			flag.DurationVar(p, om.Name, value, om.Usage)
		}
	case *float64:
		if value, ok := om.Value.(float64); !ok {
			panic(fmt.Errorf("option '%s' float64 default value not float64: %T", om.Name, om.Value))
		} else {
			flag.Float64Var(p, om.Name, value, om.Usage)
		}
	case *int64:
		if value, ok := om.Value.(int64); !ok {
			panic(fmt.Errorf("option '%s' Int64 default value not int64: %T", om.Name, om.Value))
		} else {
			flag.Int64Var(p, om.Name, value, om.Usage)
		}
	case *int:
		if value, ok := om.Value.(int); !ok {
			panic(fmt.Errorf("option '%s' Int default value not int: %T", om.Name, om.Value))
		} else {
			flag.IntVar(p, om.Name, value, om.Usage)
		}
	case *string:
		if value, ok := om.Value.(string); !ok {
			panic(fmt.Errorf("option '%s' String default value not string: %T", om.Name, om.Value))
		} else {
			flag.StringVar(p, om.Name, value, om.Usage)
		}
	case *uint64:
		if value, ok := om.Value.(uint64); !ok {
			panic(fmt.Errorf("option '%s' Uint64 default value not uint64: %T", om.Name, om.Value))
		} else {
			flag.Uint64Var(p, om.Name, value, om.Usage)
		}
	case *uint:
		if value, ok := om.Value.(uint); !ok {
			panic(fmt.Errorf("option '%s' Uint default value not uint: %T", om.Name, om.Value))
		} else {
			flag.UintVar(p, om.Name, value, om.Usage)
		}
	case *[]string:
		if value, ok := om.Value.([]string); !ok {
			panic(fmt.Errorf("option '%s' []string default value not []string: %T", om.Name, om.Value))
		} else {
			flag.Var(GetStringSliceValue(p, value), om.Name, om.Usage)
		}
	default:
		panic(fmt.Errorf("option '%s' Unknown options type: %T", om.Name, p))
	}
}

const (
	typeMismatch = "option %s type %T: bad yaml data pointer of type: %T"
)

func (om *OptionData) ApplyYaml() (err error) {
	if om.Y == nil {
		return // does not have YamlValue
	}
	yamlDataPointer := om.Y.Pointer
	parl.Debug("optionData.ApplyYaml option: -%s yaml key:%q", om.Name, om.Y.Name)
	switch valuePointer := om.P.(type) {
	case *bool:
		typedPointer, ok := yamlDataPointer.(*bool)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *time.Duration:
		typedPointer, ok := yamlDataPointer.(*time.Duration)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *float64:
		typedPointer, ok := yamlDataPointer.(*float64)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *int64:
		typedPointer, ok := yamlDataPointer.(*int64)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *int:
		typedPointer, ok := yamlDataPointer.(*int)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *string:
		typedPointer, ok := yamlDataPointer.(*string)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *uint64:
		typedPointer, ok := yamlDataPointer.(*uint64)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *uint:
		typedPointer, ok := yamlDataPointer.(*uint)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *[]string:
		typedPointer, ok := yamlDataPointer.(*[]string)
		if !ok {
			return perrors.Errorf(typeMismatch, om.Name, om.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	default:
		return perrors.Errorf("option %s: unknown value type: %T", om.Name, om.P)
	}
	return
}
