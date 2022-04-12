/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

type StatuserFactory interface {
	NewStatuser(useStatuser bool, d time.Duration) (statuser Statuser)
}

/*
Statuser prints threads that do not react within a specified time frame.
This is useful during early multi-threaded design when it is uncertain why work
is not progressing. Is it your code or their code?
Once the program concludes without hung threads,
Tracer is the better tool to identify issues.
*/
type Statuser interface {
	Set(status string) (statuser Statuser)
	Shutdown()
}
