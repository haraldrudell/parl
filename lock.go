/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
)

// parl.Mutex is a one-liner mutex
//   - sync.Mutex: Lock() TryLock() Unlock()
//
// usage:
//
//	defer m.Lock().Unlock()
type Mutex struct{ sync.Mutex }

// Lock returns lock reference
func (m *Mutex) Lock() (m2 Unlocker) { m.Mutex.Lock(); return m }
