//go:build !android && !illumos && !ios && !js && !plan9 && !wasip1 && !windows

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"os"
	"testing"

	"github.com/haraldrudell/parl"
	"golang.org/x/sys/unix"
)

func TestSignalString(t *testing.T) {
	//t.Error("logging on")
	var osSignalNil os.Signal
	var osSignalM1 os.Signal = unix.Signal(-1)
	var osSignal0 os.Signal = unix.Signal(0)
	const noPanic = false
	const isPanic = true
	var osSignal os.Signal = unix.SIGSEGV
	// os.Signal type: syscall.Signal String: "segmentation fault" unix.SIGSEGV type: syscall.Signal
	t.Logf("os.Signal type: %T String: %[1]q unix.SIGSEGV type: %T", osSignal, unix.SIGSEGV)
	type args struct {
		signal os.Signal
	}
	tests := []struct {
		name    string
		args    args
		wantS   string
		isPanic bool
	}{
		{"nil os.Signal", args{osSignalNil}, "", isPanic},
		{"SIGSEGV", args{unix.SIGSEGV}, "signal “segmentation fault” SIGSEGV 11 0xb", noPanic},
		{"-1", args{osSignalM1}, "signal “signal -1” -1 -0x1", noPanic},
		{"0", args{osSignal0}, "signal “signal 0” 0 0x0", noPanic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotS, isPanic = invokeSignalString(tt.args.signal)
			if isPanic && tt.isPanic {
				return // expected panic, got panic
			} else if isPanic && !tt.isPanic {
				t.Errorf("PANIC SignalString() = %v, want %v", gotS, tt.wantS)
				return
			} else if !isPanic && tt.isPanic {
				t.Errorf("NOPANIC %s", tt.name)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("SignalString() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

func invokeSignalString(signal os.Signal) (s string, isPanic bool) {
	var err error
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err, &isPanic)

	s = SignalString(signal)
	return
}
