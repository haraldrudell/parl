/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

const (
	// counts public CachedLocation method, [CachedLocation.get], [CacheMechanic.EnsureInit]
	// [CachedLocation.init]
	cachedLocationFrames = 4
)

// CachedLocation caches the code location and performantly provides it in string formats. Thread-safe
//   - one top-level CachedLocation variable is required for each code location
//   - code line provided is the location of first getter method
//   - caching saves 1,003 ns, ie. 0.85 parallel mutex Lock/Unlock
//   - cannot use sync.Once.Do because number of frames it consumes is unknown
//   - initialization-free, thread-safe
//
// usage:
//
//	var mycl pruntime.CachedLocation
//	func f() {
//	  println(mycl.PackFunc())
type CachedLocation struct {
	m CacheMechanic
	// payload values written inside lock
	packFunc, short, funcName, funcIdentifier string
}

// "mains.AddErr" Thread-safe
//   - similar to [perrors.NewPF] or [perrors.ErrorfPF]
func (c *CachedLocation) PackFunc() (packFunc string) { return c.get().packFunc }

// "myFunc" Thread-safe
func (c *CachedLocation) FuncIdentifier() (funcIdentifier string) { return c.get().funcIdentifier }

// "mains.(*Executable).AddErr-executable.go:25" Thread-safe
//   - similar to [perrors.Short] location
func (c *CachedLocation) Short() (location string) { return c.get().short }

// "github.com/haraldrudell/parl/mains.(*Executable).AddErr" Thread-safe
//   - FuncName is the value compared to by [parl.SetRegexp]
func (c *CachedLocation) FuncName() (location string) { return c.get().funcName }

// get ensures data is loaded exactly once
func (c *CachedLocation) get() (c2 *CachedLocation) {
	c.m.EnsureInit(c.init)
	return c
}

// init is invoked on the very first invocation inside lock
func (c *CachedLocation) init() {
	var codeLocation = NewCodeLocation(cachedLocationFrames)
	c.packFunc = codeLocation.PackFunc()
	c.short = codeLocation.Short()
	c.funcName = codeLocation.FuncName
	c.funcIdentifier = codeLocation.FuncIdentifier()
}
