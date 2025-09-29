/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package mains

// BaseOptionsType is the type that holds mains’ effective option values
//   - -yamlFile -yamlKey -verbose -debug -silent -version -no-yaml
type BaseOptionsType = struct {
	YamlFile, YamlKey, Verbosity   string
	Debug, Silent, Version, DoYaml bool
}
