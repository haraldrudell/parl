/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"reflect"
	"testing"
)

func TestErrorData(t *testing.T) {
	//t.Errorf("Logging on")
	var key, value = "key", "value"
	var message = "error-message"
	var eData *errorData
	var typeName = reflect.TypeOf(eData).String()
	var m = map[CSFormat]string{
		DefaultFormat: message,
		CodeLocation:  "",
		ShortFormat:   message,
		LongFormat:    message + " [" + typeName + "]\n" + key + ": " + value,
		ShortSuffix:   "",
		LongSuffix:    key + ": " + value,
	}

	var formatAct, formatExp, keyAct, valueAct string
	var ok bool
	var err error
	var err0 = errors.New(message)
	var hasData ErrorHasData

	// ChainString() KeyValue()
	var _ *errorData

	err = NewErrorData(err0, key, value)

	// err should be ErrorHasData
	hasData, ok = err.(ErrorHasData)
	if !ok {
		t.Fatalf("NewErrorData not ErrorHasData")
	}

	// err should be eData
	eData, ok = err.(*errorData)
	if !ok {
		t.Fatalf("NewErrorData not ErrorHasData")
	}

	// KeyValue should match
	keyAct, valueAct = hasData.KeyValue()
	if keyAct != key {
		t.Errorf("key %q exp %q", keyAct, key)
	}
	if valueAct != value {
		t.Errorf("value %q exp %q", valueAct, value)
	}

	for _, csFormat := range csFormatList {
		if formatExp, ok = m[csFormat]; !ok {
			t.Errorf("no expected value for format: %s", csFormat)
		}
		formatAct = eData.ChainString(csFormat)

		// DefaultFormat: "error-message"
		// CodeLocation: ""
		// ShortFormat: "error-message"
		//  LongFormat: "error-message [*errorglue.errorData]\nkey: value"
		// ShortSuffix: ""
		// LongSuffix: "key: value"
		t.Logf("%s: %q", csFormat, formatAct)

		// ChainString should match
		if formatAct != formatExp {
			t.Errorf("FAIL %s:\n%q exp\n%q",
				csFormat, formatAct, formatExp,
			)
		}
	}
}
