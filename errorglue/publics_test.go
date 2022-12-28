/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestErrorChainSlice(t *testing.T) {
	err := errors.New("x")
	err2 := fmt.Errorf("%w", err)
	errLen := 2

	var errs []error

	errs = ErrorChainSlice(nil)
	if len(errs) != 0 {
		t.Errorf("errs len not 0: %d", len(errs))
	}

	errs = ErrorChainSlice(err2)
	if len(errs) != errLen {
		t.Errorf("errs len not %d: %d", errLen, len(errs))
	}
}

func TestGetInnerMostStack(t *testing.T) {
	err := errors.New("x")
	errOneStack := NewErrorStack(err, pruntime.NewStackSlice(0))
	stack1 := errOneStack.(ErrorCallStacker).StackTrace()
	errTwoStacks := NewErrorStack(errOneStack, pruntime.NewStackSlice(1))
	stack2 := errTwoStacks.(ErrorCallStacker).StackTrace()
	if len(stack1) == len(stack2) {
		t.Error("stacks same")
	}

	var stack pruntime.StackSlice

	stack = GetInnerMostStack(nil)
	if len(stack) != 0 {
		t.Errorf("stack1 len not 0: %d", len(stack))
	}

	stack = GetInnerMostStack(err)
	if len(stack) != 0 {
		t.Errorf("stack2 len not 0: %d", len(stack))
	}

	stack = GetInnerMostStack(errOneStack)
	if len(stack) == 0 {
		t.Error("stack3 len 0")
	}
	if len(stack) != len(stack1) {
		t.Errorf("stack3 bad length %d exp %d", len(stack), len(stack1))
	}

	stack = GetInnerMostStack(errTwoStacks)
	if len(stack) == 0 {
		t.Error("stack4 len 0")
	}
	if len(stack) != len(stack1) {
		t.Errorf("stack4 bad length %d exp %d", len(stack), len(stack1))
	}
}

func TestGetStackTrace(t *testing.T) {
	err := errors.New("x")
	errOneStack := NewErrorStack(err, pruntime.NewStackSlice(0))
	stack1 := errOneStack.(ErrorCallStacker).StackTrace()
	errTwoStacks := NewErrorStack(errOneStack, pruntime.NewStackSlice(1))
	stack2 := errTwoStacks.(ErrorCallStacker).StackTrace()
	if len(stack1) == len(stack2) {
		t.Error("stacks same")
	}

	var stack pruntime.StackSlice

	stack = GetStackTrace(nil)
	if len(stack) != 0 {
		t.Errorf("stack1 len not 0: %d", len(stack))
	}

	stack = GetStackTrace(err)
	if len(stack) != 0 {
		t.Errorf("stack2 len not 0: %d", len(stack))
	}

	stack = GetStackTrace(errOneStack)
	if len(stack) == 0 {
		t.Error("stack3 len 0")
	}
	if len(stack) != len(stack1) {
		t.Errorf("stack3 bad length %d exp %d", len(stack), len(stack1))
	}

	stack = GetStackTrace(errTwoStacks)
	if len(stack) == 0 {
		t.Error("stack4 len 0")
	}
	if len(stack) != len(stack2) {
		t.Errorf("stack4 bad length %d exp %d", len(stack), len(stack2))
	}
}

func TestDumpChain(t *testing.T) {
	var err = errors.New("x")
	errType := fmt.Sprintf("%T", err)
	var err2 = fmt.Errorf("%w", err)
	err2Type := fmt.Sprintf("%T", err2)

	type args struct {
		err error
	}
	tests := []struct {
		name          string
		args          args
		wantTypeNames string
	}{
		{"nil", args{nil}, ""},
		{"1", args{err}, errType},
		{"2", args{err2}, err2Type + "\x20" + errType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTypeNames := DumpChain(tt.args.err); gotTypeNames != tt.wantTypeNames {
				t.Errorf("DumpChain() = %v, want %v", gotTypeNames, tt.wantTypeNames)
			}
		})
	}
}

func TestDumpGo(t *testing.T) {
	var err = errors.New("x")
	errType := fmt.Sprintf("%T %#[1]v", err)
	var err2 = fmt.Errorf("%w", err)
	err2Type := fmt.Sprintf("%T %#[1]v", err2)

	type args struct {
		err error
	}
	tests := []struct {
		name          string
		args          args
		wantTypeNames string
	}{
		{"nil", args{nil}, ""},
		{"1", args{err}, errType},
		{"2", args{err2}, err2Type + "\n" + errType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTypeNames := DumpGo(tt.args.err); gotTypeNames != tt.wantTypeNames {
				t.Errorf("DumpGo() = %v, want %v", gotTypeNames, tt.wantTypeNames)
			}
		})
	}
}
