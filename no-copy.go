/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

func (*noCopy) Lock() {}
