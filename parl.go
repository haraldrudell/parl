/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package parl handles inter-thread communication and controls parallelism

parl has sub-packages augmenting the Go standard library:
 perrors pfs plog pnet pos pruntime psql pstrings
 psyscall pterm ptime

parl has feature packages:
 ev — handling of goroutine-based functions
 goid — unique goroutine IDs
 mains — functions for writing command-line utilities and services
 parlca — self-signed certificate authority
 progress — monitor work progress for a large number of threads
 sqlite — local SQL database
 threadprof — profiling and counters for what threads are doing
 // statuser: thread hang detector
 tracer — event lists by task rather than by time or thread

parl features per-writer thread-safe logging with topic and per-package
output control:

Logging is to stderr except for the Out function.
parl logging uses comma separator for numbers.
One argument is output as string, two or more arguments is Printf.
The location matched against the  regular expression is
full package path, optional type receiver and the funtion name:
“github.com/haraldrudell/mypackage.(*MyType).MyFunc”
 Out(string, ...interface{}) — Standard output
 Log(string, ...interface{}) — Always outputs to stderr
 parl.D(string, ...interface{}) — Same as Log, intended for temporary use

 Info(string, ...interface{}) — Informational progress messages
 SetSilent(true) — removes Info output
 IsSilent() — deteremines if Info printing applies

 Debug(string, ...interface{}) — only prints for locations where SetDebug(true)
 SetDebug(true) — Control Debug() globally, code location for all prints, long stack traces
 SetRegexp(regExp string) (err error) — Regular expression controlling local Debug() printing
 IsThisDebug() — Determines if debug is active for the executing function

 Console(string, ...interface{}) — terminal interactivity output

parl provides panic recovery for process and goroutines:
capturing panics, annotating, retrieving and storing errors and
invoking error handling functions:
 func myThread(errCh chan<- error) {
   var err error
   defer close(errCh)
   defer parl.Recover2(parl.Annotation(), &err, func (err error) { errCh <- err})

   if err = someFunc(); err != nil {
     err = perrors.Errorf("someFunc: %w", err)
     return
   …

 func myThreadSafeThread(wg *sync.WaitGroup, errs *perrors.ParlError) { // ParlError: thread-safe error store
   defer wg.Done()
   defer parl.Recover(parl.Annotation(), nil, errs.AddErrorProc)

   if err = someFunc(); err != nil {
     errs.AddError(perrors.Errorf("someFunc: %w", err))
     return
   …

parl package features:
 AtomicBool — Thread-safe boolean
 Closer — Deferable, panic free channel close
 ClosableChan — Initialization-free channel with inspectable close that does not panic
 Moderator — A ticketing system for limited parallelism
 NBChan — A non-blocking unbuffered channel with trillion-size buffer
 OnceChan — Initialization-free inspectable shutdown semaphore implementing Context
 SerialDo — Serialization of invocations
 WaitGroup —Inspectable WaitGroup
 Debouncer — Invocation debouncer, pre-generics
 Sprintf — Supporting thousands separator

Parl is about 9,000 lines of Go code with first line written on November 21, 2018

On March 16th, 2022, parl was open-sourced under an ISC License

© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
*/
package parl

const (
	Rfc3339s   = "2006-01-02 15:04:05-07:00"
	Rfc3339ms  = "2006-01-02 15:04:05.999-07:00"
	Rfc3339us  = "2006-01-02 15:04:05.999999-07:00"
	Rfc3339ns  = "2006-01-02 15:04:05.999999999-07:00"
	Rfc3339sz  = "2006-01-02T15:04:05Z"
	Rfc3339msz = "2006-01-02T15:04:05.999Z"
	Rfc3339usz = "2006-01-02T15:04:05.999999Z"
	Rfc3339nsz = "2006-01-02T15:04:05.999999999Z"
)

type Password interface {
	HasPassword() (hasPassword bool)
	Password() (password string)
}

type FSLocation interface {
	Directory() (directory string)
}
