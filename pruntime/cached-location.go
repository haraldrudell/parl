/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"sync"
)

const (
	cachedLocationFrames = 2
)

// CachedLocation caches the code location and performantly provides it in string formats
//   - one top-level CachedLocation variable is required for each code location
//
// usage:
//
//	var mycl pruntime.CachedLocation
//	func f() {
//	  println(mycl.PackFunc())
type CachedLocation struct {
	initializeOnce            sync.Once
	packFunc, short, funcName string
}

// init returns a cached provider of code location in string formats
func (c *CachedLocation) init() {
	var codeLocation = NewCodeLocation(cachedLocationFrames)
	c.packFunc = codeLocation.PackFunc()
	c.short = codeLocation.Short()
	c.funcName = codeLocation.FuncName
}

// "mains.AddErr" Thread-safe
//   - similar to [perrors.NewPF] or [perrors.ErrorfPF]
func (c *CachedLocation) PackFunc() (packFunc string) {
	c.initializeOnce.Do(c.init)
	return c.packFunc
}

// "mains.(*Executable).AddErr-executable.go:25" Thread-safe
//   - similar to [perrors.Short] location
func (c *CachedLocation) Short() (location string) {
	c.initializeOnce.Do(c.init)
	return c.short
}

// "github.com/haraldrudell/parl/mains.(*Executable).AddErr" Thread-safe
//   - FuncName is the value compared to by [parl.SetRegexp]
func (c *CachedLocation) FuncName() (location string) {
	c.initializeOnce.Do(c.init)
	return c.funcName
}
