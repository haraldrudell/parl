/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"fmt"
	"strings"
)

const (
	// the metaMarker where string is put
	metaMarker = ""
	// meta formatting string: “01n02e03w004L05N06”
	//	- 01 displayLineCount
	//	- n02 number of newlines in status
	//	- e03 number of empty lines
	//	- w04 width
	//	- L05 number of long lines
	//	- N06 number of counted newlines
	metaFormat = metaMarker + "%02dn%02de%02dw%03dL%02dN%02d"
)

// - to activate d, use option “-verbose StatusTerminal..Status”
type StatusDebug struct {
	// d.n indicates StatusDebug is active
	//	- n is true if the new function was invoked
	n bool
	// displayCount copy
	metaDisplayCount int
	// number of newlines in status
	metaNewlineCount int
	// number of empty lines at end
	metaEmptyLineCount int
	// number of lines longer than width
	metaLongLines int
	// number of lines added due to newlines in status
	metaCountedNewlines int
	// current window reported width in characters
	width int
}

// - to activate d, use option “-verbose StatusTerminal..Status”
func NewStatusDebug(lines []string, width int, fieldp ...*StatusDebug) (debug *StatusDebug) {

	if len(fieldp) > 0 {
		debug = fieldp[0]
	}
	if debug == nil {
		debug = &StatusDebug{}
	}

	*debug = StatusDebug{
		n:                true,
		metaNewlineCount: len(lines) - 1,
		width:            width,
	}
	return
}

// meta formatting string: “01n02e03w004L05N06”
func (d *StatusDebug) DebugText() (text string) {
	return fmt.Sprintf(
		metaFormat, d.metaDisplayCount, d.metaNewlineCount,
		d.metaEmptyLineCount, d.width,
		d.metaLongLines, d.metaCountedNewlines,
	)
}

// inserts the process-complete status at end of status
// - without changing the length of status
func (d *StatusDebug) UpdateOutput(output string, displayLineCount int) (o string) {
	d.metaDisplayCount = displayLineCount
	var metaIndex = strings.Index(output, metaMarker)
	var metaInsert = d.DebugText()
	o = output[:metaIndex] + metaInsert + output[metaIndex+len(metaInsert):]
	return
}
