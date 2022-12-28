/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plogger

import (
	"io"
	"log"
	"os"
)

var logMap = map[io.Writer]*log.Logger{
	os.Stdout: log.New(os.Stdout, "", 0),
	os.Stderr: log.New(os.Stderr, "", 0),
}

// GetLog returns shared log.Logger instances for file descriptors os.Stdout and os.Stderr.
//   - if Logger instance is shared, output is thread-safe.
//   - if other means of output is used, the result is unpredictably intermingled output
func GetLog(writer io.Writer) *log.Logger {
	return logMap[writer]
}
