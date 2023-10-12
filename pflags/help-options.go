/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pflags

// helpOptions are additional options implemented by the flag package
//   - [mains.ArgParser] invokes [flag.Parse]
//   - [flag.FlagSet] parseOne method have these as string constants in the code
var helpOptions = []string{"h", "help"}

// returns implicit help options: "h" "help"
func HelpOptions() (optionNames []string) {
	return helpOptions
}
