// © 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License
module github.com/haraldrudell/parl/psql

go 1.23.0

toolchain go1.24.0

replace github.com/haraldrudell/parl => ../../parl

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/haraldrudell/parl v0.4.194
	github.com/haraldrudell/parl/sqliter v0.4.194
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250215185904-eff6e970281f // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
	modernc.org/sqlite v1.35.0 // indirect
)
