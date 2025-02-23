/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

const (
	// as second argument to [BaseOptionData], indicates that yaml options -yamlFile -yamlKey should not be present
	YamlNo YamlOption = false
)

// type for second argument to [BaseOptionData] [YamlNo]
type YamlOption bool
