/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package progress provides printable progress reporting for multi-threaded operations
package progress

const (
	// ValueKey returns the list of retrievable string values
	ValueKey = "values"
	// LogOutput to Value provides output formatted for logging
	LogFormat = "log"
	// TerminalFormat to Value provides output with colors
	TerminalFormat = "log"
	// PlainFormat contains no control characters other than newline
	PlainFormat = "plain"
)

type Progress interface {
	Value(key string, parameter int) (value string) // parameter may be available column count for formatting
	Values(key string, parameter int) (values []string)
}
