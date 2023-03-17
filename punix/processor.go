//go:build !darwin && !linux

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

// Processor returns a human-readable string describing the hosts’s processor model
//   - model is empty string if information is not available
func Processor() (model string, err error) {}
