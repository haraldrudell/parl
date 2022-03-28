/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package errorglue contains helful declarations that are not important
package errorglue

// ChainStringer obntain s a comprehensive string representation of an error chain.
// formats used are DefaultFormat ShortFormat LongFormat ShortSuffix LongSuffix
type ChainStringer interface {
	// ChainString is used by ChainStringer() to obtain comprehensive
	// string representations of errors.
	// The argument isIgnore() is used to avoid printing cyclic error values.
	// If a ChainStringer pointer receiver gets a nil value, the empty string is returned.
	// ChainString() obtains a string representation of the errors in its chain.
	// Rich errors implement either ChainStringer or fmt.Formatter
	ChainString(format CSFormat) string
}

// Wrapper is an interface indicating error-chain capabilities.
// It is not public in errors package
type Wrapper interface {
	Unwrap() error // Unwrap returns the next error in the chain or nil
}

// ErrorHasData enrichens an error with key and value strings
type ErrorHasData interface {
	KeyValue() (key, value string)
}

// RelatedError enrichens an error with an enclosed additional error value
type RelatedError interface {
	AssociatedError() (error error)
}

// ErrorHasCode allows an error to classify itself
type ErrorHasCode interface {
	// Check if this error claims a particular Linux errno, an int
	IsErrno(errno int) (hasErrno bool)
	// ErrorCode determines if this error claims code, a string
	ErrorCode(code string) (hasCode bool)
	// ErrorCodes returns codes that this error claims, some are numeric strings mapping to an errno
	ErrorCodes(codes []string) (has []string)
}

// ErrorCallStacker enrichens an error with a stack trace of code locations
type ErrorCallStacker interface {
	StackTrace() StackSlice
}

// ErrorStore is a thread-safe store for any number of errors
type ErrorStore interface {
	AddError(err error) (e error)
	GetError() (e error)
	Error() (message string)
}
