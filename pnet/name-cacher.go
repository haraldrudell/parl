/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

const (
	// name cache should not be used in the api call
	NoCache NameCacher = iota
	// use name cache and update the cache first
	Update
	// use name cache without update. Mechanic for ongoing name cache update is required
	NoUpdate
)

// NameCacher contaisn instructions for interface-name cache read
//   - [NoCache] [Update] [NoUpdate]
type NameCacher uint8
