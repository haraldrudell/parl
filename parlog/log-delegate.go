/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlog

import (
	"io"
	"log"
	"os"
)

var logMap = map[io.Writer]*log.Logger{
	os.Stdout: log.New(os.Stdout, "", 0),
	os.Stderr: log.New(os.Stderr, "", 0),
}

func GetLog(writer io.Writer) *log.Logger {
	return logMap[writer]
}
