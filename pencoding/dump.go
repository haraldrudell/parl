/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pencoding

// Dumper prints all values in quoted format
// enabling exact comparisons
//   - the pretty-print alternative is [StringNs]
//   - similar to [fmt.Stringer]
type Dumper interface {
	// Dump prints all fields in quoted format
	// enabling exact comparisons
	Dump() (s string)
}
