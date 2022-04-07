/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import (
	"time"

	"github.com/haraldrudell/parl"
)

type RecordDo struct {
	At   time.Time
	Text string
}

func NewRecordDo(text string) (record parl.Record) {
	return &RecordDo{At: time.Now(), Text: text}
}

func (rd *RecordDo) Values() (at time.Time, text string) {
	at = rd.At
	text = rd.Text
	return
}
