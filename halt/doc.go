/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package halt detects Go runtime execution halts.

sleep is an interval used for halt detection
  - for monotonic mode using dynamic interval configuration
  - the thread tries to get interval to at lest 1 ms to lower the cpu load
  - Go garbage collector stop-the-world max pause is 1 ms.
    Typical pause is 100 μs.
    For monitoring the garbage collector, a tentative start is 100 μs.
  - [time.Sleep] is unreliable and cannot be used.
    Time.Sleep delays at least 1 ms on Linux
  - [runtime.Gosched] may delay over 1 s for 10k threads
  - [time.NewTimer] delays at least 374 ns
  - [time.NewTicker] is least troublesome sleep method.
    Fastest interval when 1 ns requested is 142.6 ns for single-thread
  - [time.Now] is 1 μs precision on macOS
  - 1 KiB allocation make([]byte, 1024) is 151.4 ns
  - context switch from one thread to another via lock is 224.7 ns
  - much be large enough so that slice allocation or thread-switch
    does not trigger a report
  - ticker, switch and allocation is 518.7 ns (142.6 + 151.4 + 224.7)
  - on macOS:
  - — smallest detectable halt is 11–29 µs
  - — 10 ms halt is exhibited within 1 s
  - — 30 ms halt is exhibited within 1 minute
  - on Linux:
  - — smallest detectable halt is 4.495 µs
  - — 2 ms halt is exhibited within 5 s
  - — 4 ms halt is exhibited within 2 minutes
*/
package halt
