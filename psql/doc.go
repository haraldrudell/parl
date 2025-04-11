/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package psql provides cached, shared, partitioned database objects with
// cached prepared statements cancelable by context.
//
//   - [NewDBMap] provides cached, shared database implementation objects.
//     database access is using cached prepared statements and
//     access using application name and partition name like year.
//   - [NewResultSetIterator] provides a Go for-statements abstract result-set iterator
//   - [ScanFunc] is the signature for preparing custom result-set iterators
//   - seamless statement-retry, remedying concurrency-deficient databases such as
//     SQLite3
//   - —
//   - [TrimSql] trims SQL statements in Go raw string literals, multi-line strings enclosed by
//     the back-tick ‘`’ character
//   - [ColumnType] describes columns of a result-set
//   - convenience method for single-value queries: [DBMap.QueryInt] [DBMap.QueryString]
//   - [SqlExec] pprovides statement execution prior to obtaining a cached database, ie. for
//     seamlessly preparing the schema
package psql
