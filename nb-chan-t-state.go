/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Thread status values
const (
	NBChanExit      NBChanTState = "exit"     // NBChan thread is not running
	NBChanAlert     NBChanTState = "alert"    // NBChan thread is always running and blocked idle waiting for alert
	NBChanGets      NBChanTState = "GetWait"  // NBChan thread is blocked waiting for Get invocations to complete
	NBChanSends     NBChanTState = "SendWait" // NBChan thread is blocked waiting for Send/SendMany invocations to complete
	NBChanSendBlock NBChanTState = "chSend"   // NBChan thread is blocked in channel send
	NBChanRunning   NBChanTState = "run"      // NBChan is running
	NBChanNoLaunch  NBChanTState = "none"     // thread was never launched
)

// state of NBChan thread
//   - NBChanExit NBChanAlert NBChanGets NBChanSends NBChanSendBlock
//     NBChanRunning
type NBChanTState string
