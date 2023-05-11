/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"sync"
	"sync/atomic"
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
	initLock                                  sync.Mutex
	isReady                                   atomic.Bool // written inside lock, read provides thread-safety
	packFunc, short, funcName, funcIdentifier string      // written inside lock
}

// init returns a cached provider of code location in string formats
func (c *CachedLocation) init() {
	c.initLock.Lock()
	defer c.initLock.Unlock()

	if c.isReady.Load() {
		return // was already set
	}
	var codeLocation = NewCodeLocation(cachedLocationFrames)
	c.packFunc = codeLocation.PackFunc()
	c.short = codeLocation.Short()
	c.funcName = codeLocation.FuncName
	c.funcIdentifier = codeLocation.FuncIdentifier()
	c.isReady.Store(true) // last write to avoid race condition
}

// "mains.AddErr" Thread-safe
//   - similar to [perrors.NewPF] or [perrors.ErrorfPF]
func (c *CachedLocation) PackFunc() (packFunc string) {
	if !c.isReady.Load() {
		c.init()
	}
	return c.packFunc
}

// "myFunc"
func (c *CachedLocation) FuncIdentifier() (funcIdentifier string) {
	return c.funcIdentifier
}

// "mains.(*Executable).AddErr-executable.go:25" Thread-safe
//   - similar to [perrors.Short] location
func (c *CachedLocation) Short() (location string) {
	if !c.isReady.Load() {
		c.init()
	}
	return c.short
}

// "github.com/haraldrudell/parl/mains.(*Executable).AddErr" Thread-safe
//   - FuncName is the value compared to by [parl.SetRegexp]
func (c *CachedLocation) FuncName() (location string) {
	if !c.isReady.Load() {
		c.init()
	}
	return c.funcName
}
