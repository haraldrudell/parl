/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

const (
	// foregorund color red
	Red = "\x1b[38:5:9m"
	// foregorund color green
	Green = "\x1b[38:5:2m"
	// foreground color reset to default
	ResetColors = "\x1b[39;49m"
)

// ANSI escape sequences
const (
	// ANSI erase to end-of-line
	EraseEndOfLine = "\x1b[K"
	// ANSI erase to end-of-display
	EraseEndOfDisplay = "\x1b[J"
	// ANSI move cursor to column zero
	MoveCursorToColumnZero = "\r"
	// ANSI move cursor up
	//	- at top line does nothing
	CursorUp = "\x1b[A"
	// ANSI move cursor down to the left
	//	- on last line, causes the view to scroll
	NewLine = "\n"
	// Space for clearing old status characters
	Space = "\x20"
)
