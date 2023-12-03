/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

// OnError is a function that receives error values from an errp error pointer or a panic
type OnError func(err error)

// NoOnError is used with Recover and Recover2 to silence the default error logging
func NoOnError(err error) {}
