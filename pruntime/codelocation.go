/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	newCodeLocationStackFrames = 1
)

// CodeLocation is similar to runtime.Frame, but contains basic types
// string and int only
type CodeLocation struct {
	// File is the absolute path to the go source file
	//  /opt/foxyboy/sw/privates/parl/mains/executable.go
	File string
	// Line is the line number in the source file
	//  35
	Line int
	// FuncName is the fully qualified Go package path,
	// a possible value or pointer receiver struct name,
	// and the function name
	//  github.com/haraldrudell/parl/mains.(*Executable).AddErr
	FuncName string
}

// NewCodeLocation gets data for a single stack frame.
// if stackFramesToSkip it returns data for the caller of NewCodeLocation
func NewCodeLocation(stackFramesToSkip int) (cl *CodeLocation) {
	if stackFramesToSkip < 0 {
		stackFramesToSkip = 0
	}
	c := CodeLocation{}

	var pc uintptr
	var ok bool
	// pc: opaque
	// file: basename.go
	// line: int 25
	if pc, c.File, c.Line, ok = runtime.Caller(newCodeLocationStackFrames + stackFramesToSkip); !ok {
		panic(errors.New("runtime.Caller failed"))
	}
	// rFunc: runtime.Func is all opaque fields. methods:
	// Entry() (uintptr)
	// FileLine(uintptr) (line string, line int) "/opt/foxyboy/sw/privates/parl/mains/executable.go"
	// Name(): github.com/haraldrudell/parl/mains.(*Executable).AddErr
	if rFunc := runtime.FuncForPC(pc); rFunc != nil {
		c.FuncName = rFunc.Name()
	}
	return &c
}

// GetCodeLocation converts a runtime stack frame to a code location
// stack frame.
// runtime contains opaque types while code location consists of basic
// types int and string only
func GetCodeLocation(rFrame *runtime.Frame) (cl *CodeLocation) {
	// runtime.Frame:
	// PC uintptr, Func *Func, Function string, File string, Line int, Entry uintptr, funcInfo funcInfo
	return &CodeLocation{
		File:     rFrame.File,
		Line:     rFrame.Line,
		FuncName: rFrame.Function,
	}
}

// FuncName returns function name, characters no space:
//   AddErr
func (cl *CodeLocation) Name() (funcName string) {
	packageAndFunc := filepath.Base(cl.FuncName)
	if lastDotIndex := strings.LastIndex(packageAndFunc, "."); lastDotIndex >= 0 {
		return packageAndFunc[lastDotIndex+1:]
	}
	return packageAndFunc
}

// Package return base package name, a single word of characters with no space:
//   mains
func (cl *CodeLocation) Package() (funcName string) {
	packageAndFunc := filepath.Base(cl.FuncName)
	if dotIndex := strings.Index(packageAndFunc, "."); dotIndex >= 0 {
		return packageAndFunc[:dotIndex]
	}
	return packageAndFunc
}

// PackFunc return base package name and function:
//   mains.AddErr
func (cl *CodeLocation) PackFunc() (packageDotFunction string) {
	return cl.Package() + "." + cl.Name()
}

// Base returns base package name, an optional type name and the function name:
//   mains.(*Executable).AddErr
func (cl *CodeLocation) Base() (baseName string) {
	return filepath.Base(cl.FuncName)
}

// Short returns base package name, an optional type name and
// the function name, base filename and line number:
//   mains.(*Executable).AddErr-executable.go:25
func (cl *CodeLocation) Short() (funcName string) {
	return fmt.Sprintf("%s-%s:%d", filepath.Base(cl.FuncName), filepath.Base(cl.File), cl.Line)
}

// Long returns full package path, an optional type name and
// the function name, base filename and line number:
//   mains.(*Executable).AddErr-executable.go:25
func (cl *CodeLocation) Long() (funcName string) {
	return fmt.Sprintf("%s-%s:%d", cl.FuncName, filepath.Base(cl.File), cl.Line)
}

// Full returns all available information on one line
// the function name, base filename and line number:
//   mains.(*Executable).AddErr-executable.go:25
func (cl *CodeLocation) Full() (funcName string) {
	return fmt.Sprintf("%s-%s:%d", cl.FuncName, cl.File, cl.Line)
}

func (cl *CodeLocation) IsSet() (isSet bool) {
	return cl.File != "" || cl.FuncName != ""
}

// String returns a two-line string representation suitable for a multi-line stack trace.
// Typical output:
//   github.com/haraldrudell/parl/error116.(*TypeName).FuncName\n
//     /opt/sw/privates/parl/error116/codelocation_test.go:20
func (cl CodeLocation) String() string {
	return fmt.Sprintf("%s\n\x20\x20%s:%d", cl.FuncName, cl.File, cl.Line)
}
