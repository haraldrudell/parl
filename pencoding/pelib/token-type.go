/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pelib

import "fmt"

const (
	IsOther TokenType = iota
	IsEOF
	IsOpenTag
	IsCloseTag
	IsProcInst
)

type TokenType uint8

func (t TokenType) String() (s string) {
	if s = tokenTypeMap[t]; s != "" {
		return
	}
	s = fmt.Sprintf("?token:%d", t)

	return
}

var tokenTypeMap = map[TokenType]string{
	IsOther:    "token-other",
	IsEOF:      "token:EOF",
	IsOpenTag:  "token:open",
	IsCloseTag: "token:close",
	IsProcInst: "token:procInst",
}
