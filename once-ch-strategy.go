/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

const (
	// [OnceCh.IsWinner] loser threads do not wait
	NoOnceWait OnceChStrategy = true
)

// whether losers wait for winner completion [NoOnceWait]
//   - for [OnceCh.IsWinner]
type OnceChStrategy bool
