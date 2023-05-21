/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"errors"
	"strconv"
	"strings"
)

// SplitAbsoluteFunctionName splits an absolute function name into its parts
//   - input: github.com/haraldrudell/parl/error116.(*TypeName).FuncName[...]
//   - packagePath: "github.com/haraldrudell/parl/"
//   - packageName: "error116" single identifier, not empty
//   - typePath: "(*TypeName)" may be empty
//   - funcName: "FuncName[...]"
func SplitAbsoluteFunctionName(absPath string) (
	packagePath, packageName, typePath, funcName string) {

	// get multiple-slashes package path excluding single-word base package name
	// "github.com/haraldrudell/parl/"
	// "error116.(*TypeName).FuncName[...]"
	remainder := absPath
	if lastSlash := strings.LastIndex(remainder, "/"); lastSlash != -1 {
		// "github.com/haraldrudell/parl/"
		packagePath = remainder[:lastSlash+1]
		remainder = remainder[lastSlash+1:]
	}

	// get base package name: "error116"
	periodIndex := strings.Index(remainder, ".")
	if periodIndex == -1 {
		panic(errors.New("no period ending package name: " + strconv.Quote(absPath)))
	}
	packageName = remainder[:periodIndex]
	remainder = remainder[periodIndex+1:]
	if remainder == "" {
		panic(errors.New("frame ends with package name: " + strconv.Quote(absPath)))
	}

	// get types: "(*TypeName)"
	if remainder[0:1] == "(" {
		endIndex := strings.Index(remainder, ")")
		if endIndex == -1 {
			panic(errors.New("package types ')' mising: " + strconv.Quote(absPath)))
		}
		// "(*TypeName)"
		typePath = remainder[:endIndex+1]
		if endIndex+2 >= len(remainder) || remainder[endIndex+1:endIndex+2] != "." {
			panic(errors.New("no function name after ')': " + strconv.Quote(absPath)))
		}
		remainder = remainder[endIndex+2:]
	}

	// "FuncName[...]"
	funcName = remainder
	return
}
