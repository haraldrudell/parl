/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package error116 enrichens error values with string data, stack traces, associated errors,
less severe warnings, thread-safe containers and
comprehensive error string representations.

Creating errors with stack traces:
  err := error116.New("error message") // adds a stacktrace contained inside of err
  err = error116.Errorf("Some text: '%w'", err) // adds a stacktrace if err does not already have it
  err = error116.Stackn(err, 0) // adds a stack trace even if err already has it
Enriching errors:
	err = error116.AddKeyValue(err, "record", "123") // adds a key-value string, last key wins
	err = error116.AddKeyValue(err, "", "2022-03-19 11:10:00") // adds a string to a list, oldest first
Encapsulating associated errors allowing for a single error value to contain multiple errors:
  err = error116.AppendError(err, err2) // err2 is inside of err, can be printed and retrieved
  fn := error116.Errp(&err) // fn(error) can be repeatedly invoked, aggregating errors in err
  error116.ParlError.AddError(err) // thread-safe error encapsulation
Marking errors as less severe:
  warning := error116.Warning(err) // warning should not terminate the thread
  error116.IsWarning(err) // Determine severity of an error value
Printing rich errors:
  error116.Long(err)
  → error-message
  → github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
  →   /opt/sw/privates/parl/error116/chainstring_test.go:26
  → runtime.goexit
  →   /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
  → record: 1234
  → Other error: Close failed
  error116.Short(err)
  → error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
  fmt.Println(err)
  → error-message

Can be used with Printf, but only works if last error in chain is from error116:
  fmt.Printf("err: %+v", err) // same as Long()
  fmt.Printf("err: %-v", err) // save as Short()
Is compatible:
  fmt.Println(err) // no change
  fmt.Printf("err: %v", err) // no change
  err.Error() // no change
  fmt.Printf("err: %s", err) // no change
  fmt.Printf("err: %q", err) // no change

© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
*/
package perrors
