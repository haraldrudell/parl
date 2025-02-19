/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

const (
	G386     GOARCH = "386"
	Amd64    GOARCH = "amd64"
	Arm      GOARCH = "arm"
	Arm64    GOARCH = "arm64"
	RiscV64  GOARCH = "riscv64"
	Wasm     GOARCH = "wasm"
	Loong64  GOARCH = "loong64"
	Mips     GOARCH = "mips"
	Mips64   GOARCH = "mips64"
	Mips64le GOARCH = "mips64le"
	Mipsle   GOARCH = "mipsle"
	PPC64    GOARCH = "ppc64"
	PPC64le  GOARCH = "ppc64le"
	S390x    GOARCH = "s390x"
)

// GOARCH are the processor architectures supported by go1.24
//   - [G386] [Amd64] [Arm] [Arm64] [RiscV64] [Wasm] [Loong64]
//     [Mips] [Mips64] [Mips64le] [Mipsle] [PPC64] [PPC64le] [S390x]
//   - go tool dist list
type GOARCH string

func (g GOARCH) String() (s string) { return string(g) }
