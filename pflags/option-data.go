/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pflags

import (
	"flag"
	"fmt"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	typeMismatch = "option %s type %T: bad yaml data pointer of type: %T"
)

// OptionData contain options data for the flag package
//   - OptionData is used for command-line options to be declarative
//   - instead of flag.BoolVar and similar invocations scattered about the codebase
//     all options are in a single OptionData slice that is iterated so that AddOption is invoked for each element
//   - to invoke [flag.BoolVar]. pointer, name, value and usage is required
//   - [flag.BoolVar] is consumer-provided storage, [flag.Bool] is flag-provided storage
type OptionData struct {
	P     interface{} // P is a reference to where the effective value of this option is stored
	Name  string      // Name is the option name without hyphen, “debug” for -debug
	Value interface{} // Value is the default value for this option
	Usage string      // printable string describing what this option does
	Y     interface{} // reference to effective value in YamlData
}

// AddOption executes flag.BoolVar and such on the options map
//   - this defines the option to the flag package
//   - all options must be defined prior to invoking [flag.Parse]
//   - options are defined using value pointers, ie [flag.BoolVar] not flag.Bool[],
//     so [flag.Parse] updates effective option values directly
func (o *OptionData) AddOption() {
	switch effectiveValuep := o.P.(type) {
	case *bool:
		if value, ok := o.Value.(bool); !ok {
			o.defaultValuePanic()
		} else {
			flag.BoolVar(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*bool); !ok {
				o.yamlPointerPanic()
			}
		}
	case *time.Duration:
		if value, ok := o.Value.(time.Duration); !ok {
			o.defaultValuePanic()
		} else {
			flag.DurationVar(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*time.Duration); !ok {
				o.yamlPointerPanic()
			}
		}
	case *float64:
		if value, ok := o.Value.(float64); !ok {
			o.defaultValuePanic()
		} else {
			flag.Float64Var(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*float64); !ok {
				o.yamlPointerPanic()
			}
		}
	case *int64:
		if value, ok := o.Value.(int64); !ok {
			o.defaultValuePanic()
		} else {
			flag.Int64Var(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*int64); !ok {
				o.yamlPointerPanic()
			}
		}
	case *int:
		if value, ok := o.Value.(int); !ok {
			o.defaultValuePanic()
		} else {
			flag.IntVar(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*int); !ok {
				o.yamlPointerPanic()
			}
		}
	case *string:
		if value, ok := o.Value.(string); !ok {
			o.defaultValuePanic()
		} else {
			flag.StringVar(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*string); !ok {
				o.yamlPointerPanic()
			}
		}
	case *uint64:
		if value, ok := o.Value.(uint64); !ok {
			o.defaultValuePanic()
		} else {
			flag.Uint64Var(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*uint64); !ok {
				o.yamlPointerPanic()
			}
		}
	case *uint:
		if value, ok := o.Value.(uint); !ok {
			o.defaultValuePanic()
		} else {
			flag.UintVar(effectiveValuep, o.Name, value, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*uint); !ok {
				o.yamlPointerPanic()
			}
		}
	case *[]string:
		if value, ok := o.Value.([]string); !ok {
			o.defaultValuePanic()
		} else {
			flag.Var(NewStringSliceValue(effectiveValuep, value), o.Name, o.Usage)
		}
		if y := o.Y; y != nil {
			if _, ok := y.(*[]string); !ok {
				o.yamlPointerPanic()
			}
		}
	default:
		panic(perrors.ErrorfPF("option '%s' Unknown options type: %T", o.Name, effectiveValuep))
	}
}

// ApplyYaml copies a value read from yaml into the effective value location
func (o *OptionData) ApplyYaml() (err error) {

	// check if this OptionData is updated from yaml
	var yamlDataPointer = o.Y
	if yamlDataPointer == nil {
		return // does not have YamlValue
	}

	// assign the effective value with the YamlData field’s value
	switch valuePointer := o.P.(type) {
	case *bool:
		typedPointer, ok := yamlDataPointer.(*bool)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *time.Duration:
		typedPointer, ok := yamlDataPointer.(*time.Duration)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *float64:
		typedPointer, ok := yamlDataPointer.(*float64)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *int64:
		typedPointer, ok := yamlDataPointer.(*int64)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *int:
		typedPointer, ok := yamlDataPointer.(*int)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *string:
		typedPointer, ok := yamlDataPointer.(*string)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *uint64:
		typedPointer, ok := yamlDataPointer.(*uint64)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *uint:
		typedPointer, ok := yamlDataPointer.(*uint)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	case *[]string:
		typedPointer, ok := yamlDataPointer.(*[]string)
		if !ok {
			return perrors.Errorf(typeMismatch, o.Name, o.P, yamlDataPointer)
		}
		*valuePointer = *typedPointer
	default:
		return perrors.Errorf("option %s: unknown value type: %T", o.Name, o.P)
	}

	return
}

func (o *OptionData) defaultValuePanic() {
	var expTyp = fmt.Sprintf("%T", o.P)
	panic(perrors.ErrorfPF("option '%s': default value not %s: %T", o.Name, expTyp[1:], o.Value))
}

func (o *OptionData) yamlPointerPanic() {
	panic(perrors.ErrorfPF("option '%s' yaml pointer not %T: %T", o.Name, o.P, o.Y))
}

func (o *OptionData) ValueDump() (valueS string) {
	switch effectiveValuep := o.P.(type) {
	case *bool:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *time.Duration:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *float64:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *int64:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *int:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *string:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *uint64:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *uint:
		return fmt.Sprintf("%v", *effectiveValuep)
	case *[]string:
		return fmt.Sprintf("%v", *effectiveValuep)
	default:
		panic(perrors.ErrorfPF("option '%s' Unknown options type: %T", o.Name, effectiveValuep))
	}
}
