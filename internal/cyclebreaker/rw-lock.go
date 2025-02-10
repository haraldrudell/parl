/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import (
	"sync"

	"github.com/haraldrudell/parl/psync"
)

// parl.Mutex is a one-liner mutex
//   - sync.Mutex: Lock() TryLock() Unlock()
//
// usage:
//
//	defer m.Lock().Unlock()
type RWMutex struct{ sync.RWMutex }

// Lock returns lock reference
func (m *RWMutex) Lock() (m2 psync.Unlocker) { m.RWMutex.Lock(); return m }

// RLock returns lock reference
func (m *RWMutex) RLock() (m2 psync.RUnlocker) { m.RWMutex.RLock(); return m }
