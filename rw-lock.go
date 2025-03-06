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
type RWMutex struct{ sync.RWMutex }

// Lock returns lock reference
func (m *RWMutex) Lock() (m2 Unlocker) { m.RWMutex.Lock(); return m }

// RLock returns lock reference
func (m *RWMutex) RLock() (m2 RUnlocker) { m.RWMutex.RLock(); return m }
