/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import (
	"time"

	"github.com/haraldrudell/parl"
)

type TracerRecord struct {
	At   time.Time
	Text string
}

func NewTracerRecord(text string) (record parl.TracerRecord) {
	return &TracerRecord{At: time.Now(), Text: text}
}

func (rd *TracerRecord) Values() (at time.Time, text string) {
	at = rd.At
	text = rd.Text
	return
}
