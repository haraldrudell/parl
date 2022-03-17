/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"
	"log"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var logMap = map[io.Writer]*log.Logger{
	os.Stdout: log.New(os.Stdout, "", 0),
	os.Stderr: log.New(os.Stderr, "", 0),
}
var stdoutLog = logMap[os.Stdout]
var stdoutOutput = stdoutLog.Output
var stderrLog = logMap[os.Stdout]

var mSprintf = message.NewPrinter(language.English).Sprintf

func sprintf(format string, a ...interface{}) string {
	if len(a) > 0 {
		format = mSprintf(format, a...)
	}
	return format
}
