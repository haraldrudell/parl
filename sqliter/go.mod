// © 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/sqliter

go 1.23.1

toolchain go1.24.0

replace github.com/haraldrudell/parl => ../../parl

replace github.com/haraldrudell/parl/psql => ../psql

require (
	github.com/google/uuid v1.6.0
	github.com/haraldrudell/parl v0.4.208
	github.com/haraldrudell/parl/psql v0.4.208
	modernc.org/sqlite v1.36.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.9.1 // indirect
)
