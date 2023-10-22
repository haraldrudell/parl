/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"io"
	"log"
	"os"
)

// logMap translates an [io.Writer] instance to a [log.logger] instance
//   - key: io.Writer such as os.Stderr
//   - value: log.Logger obtaiend from [log.New]
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
