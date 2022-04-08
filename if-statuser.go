/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

type StatuserFactory interface {
	NewStatuser(useStatuser bool, d time.Duration) (statuser Statuser)
}

type Statuser interface {
	Set(status string) (statuser Statuser)
	Shutdown()
}
