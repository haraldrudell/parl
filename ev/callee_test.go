/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ev

import (
	"testing"

	"github.com/google/uuid"
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/evx"
)

func TestResultPanic(t *testing.T) {

	// test recover from panic
	doFn("PANIC", "non-error value: int 1", func(ctx Callee) (err error) {
		defer ctx.Result(&err)
		panic(1)
	}, t)

	// test recover from error
	errX := "errX"
	doFn("ERR", errX, func(ctx Callee) (err error) {
		defer ctx.Result(&err)
		return parl.New(errX)
	}, t)

	// test success
	doFn("SUCCESS", "", func(ctx Callee) (err error) {
		defer ctx.Result(&err)
		return
	}, t)

}

func doFn(testID, expectedErr string, fn func(ctx Callee) (err error), t *testing.T) {
	// values
	name := "name"

	// context
	gID := uuid.New()
	eventCh := make(chan Event, 3)
	ctx := NewCallee(name, GoID(gID), eventCh, nil)

	// variables
	var errValue error
	var event Event
	var actualInt int
	var actualBool bool
	var actual string

	// invoke test function
	errValue = fn(ctx)

	if expectedErr == "" {
		if errValue != nil {
			t.Logf("%s ctx.Result set err: %v", testID, errValue)
			t.Fail()
		}
	} else {
		if errValue == nil {
			t.Logf("%s ctx.Result did not set err", testID)
			t.FailNow()
		}
		actual = errValue.Error()
		if actual != expectedErr {
			t.Logf("%s ctx.Result err: expected: %q actual: %s", testID, expectedErr, errValue)
			t.FailNow()
		}
	}
	actualInt = len(eventCh)
	if actualInt != 1 {
		t.Logf("%s ctx.Result did not send 1 message on panic: %d", testID, actualInt)
		t.FailNow()
	}
	if event, actualBool = <-eventCh; !actualBool {
		t.Logf("%s Read event channel failed", testID)
		t.FailNow()
	}
	switch e := event.(type) {
	case *ExitEvent:
		if e.Err != errValue {
			t.Logf("%s ExitEvent bad Err: expected: %v actual: %v", testID, errValue, e.Err)
			t.Fail()
		}
	default:
		t.Logf("%s Unknown event type: %T", testID, event)
		t.FailNow()
	}
}

func TestSend(t *testing.T) {
	value := 3
	type DataType int
	var event Event

	mgr := NewManager(nil)
	closer := make(chan struct{})

	// the goroutine
	go func(ctx Callee) {
		defer close(closer)
		defer ctx.Result(nil)
		ctx.Send(DataType(value))
	}(mgr.CalleeContext())

	// get the listen event containing the address
	getEvent := func() (event Event, isEnd bool) {
		select {
		case <-closer:
			isEnd = true
			return
		case event = <-mgr.Events():
		}
		t.Logf("Event: %T %[1]v", event)
		if err := mgr.ProcessEvent(event); err != nil {
			t.Errorf("Bad event: %T %[1]v", event)
			t.FailNow()
		}
		if exitEv, ok := event.(*ExitEvent); ok {
			if exitEv.Err != nil {
				t.Errorf("Failure: %T %[1]v", event)
				t.FailNow()
			}
		}
		if dataEv, ok := event.(*DataEvent); ok {
			if warn, ok := dataEv.Payload().(evx.Warning); ok {
				t.Errorf("Warning: %T %[1]v", warn)
				t.FailNow()
			}
		}
		return
	}

	event, _ = getEvent()
	if dataEvent, ok := event.(*DataEvent); !ok {
		t.Errorf("Received other than DataEvent: %T %[1]v", event)
		return
	} else if data, ok := dataEvent.Payload().(DataType); !ok {
		t.Errorf("Payload not ListenEvent: %T", dataEvent.Payload())
		return
	} else {
		t.Logf("Value: %T %[1]v", data)
	}
}
