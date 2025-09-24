/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package sqliter provides partitioning, cached prepared statements and data conversion for SQLite3.
//
//   - SQLite3 databases are semlessly created by
//   - — using [DSNrFactory].DataSourceNamer with an application name and
//   - — executing queries from [github.com/haraldrudell/parl/psql.DBFactory].NewDB cached DB objects and
//   - — a partition ID
//   - statement-retry remedy for SQLite3 concurrency shortcomings
//   - — concurrency errors are not returned, there is a queueing implementation for retries
//   - — this convenience is semalessly provided by [github.com/haraldrudell/parl/psql.NewDBMap]
//
// conversions for:
//   - — Go time.Time to SQLite3 TEXT strict iso8601 UTC time-zone, nanosecond precision
//     [TimeToDB] [TimeToDBNullable] [ToTime] [NullableToTime]
//   - — Go bool to SQLite3 INTEGER [BoolToDB] [ToBool]
//   - — Go time.Time to SQLite3’s DATETIME TEXT iso8601-like format millisecond precision
//     [TimeToDATETIME] [DATETIMEtoTime]
//   - — Go uuid.UUID to SQLite3 TEXT [UUIDToDB] [UUIDToDB]
//
// additionally:
//   - retrieval of pragma, ie. SQLite3 performance-related database configuration
//     [Pragma] [PragmaDB] [PragmaSQL]
//   - data-source namer for application name and year partition
//     [DSNrFactory].DataSourceNamer [OpenDataSourceNamer]
//   - data-source that do not create database-files [OpenDataSourceNamerRO] for querying existing
//     databases
//   - retrieval of actionable SQLite3 error codes [Code]
package sqliter
