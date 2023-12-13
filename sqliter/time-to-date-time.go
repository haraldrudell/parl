/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// value in database: “2019-10-19 22:26:34.000 +00:00”
	msTimeString = "2006-01-02 15:04:05.000 -07:00"
	// length of SQLite TEXT DATETIME
	msTimeStringLength = len(msTimeString)
)

// echo $(date --rfc-3339=seconds) $(hostname --short) $(id --user --name) $(pwd)
// 2023-11-22 11:37:40-08:00 c68z foxyboy /home/foxyboy
// sqlite3 ~/.local/share/movy/movy.db .schema
// CREATE TABLE `Videos` (`id` UUID PRIMARY KEY, `description` VARCHAR(255),
// `start` DATETIME NOT NULL,
// `dur` INTEGER, `createdAt` DATETIME NOT NULL, `updatedAt` DATETIME NOT NULL);
// echo "($(sqlite3 ~/.local/share/movy/movy.db "SELECT start from Videos LIMIT 1"))"
// (2019-01-01 21:30:42.000 +00:00)

// the field videos.start is stored as: “2019-01-01 21:30:42.000 +00:00”
//   - SQLite [DATETIME] TEXT as ISO8601 strings “YYYY-MM-DD HH:MM:SS.SSS”
//   - time zone UTC, millisecond resolution
//   - start column is NOT NULL, ie. string will never be nil (sql.NullString)
//   - — SQLite storage classes: NULL INTEGER REAL TEXT BLOB
//
// [DATETIME]: https://sqlite.org/datatype3.html#date_and_time_datatype

// TimeToDATETIME converts Go time.Time to SQLite [DATETIME] UTC millisecond precision
//   - “2019-01-01 21:30:42.000 +00:00”
//
// [DATETIME]: https://sqlite.org/datatype3.html#date_and_time_datatype
func TimeToDATETIME(t time.Time) (dateTime string) {
	return t.UTC().Format(msTimeString)
}

// DATETIMEtoTime converts SQLite DATETIME to Go time.Time with location time.Local
//   - database stores as TEXT in UTC, millisecond precision
//   - “2019-01-01 21:30:42.000 +00:00”
func DATETIMEtoTime(sqliteText string) (t time.Time, err error) {

	// check length “2019-01-01 21:30:42.000 +00:00”
	if len(sqliteText) != msTimeStringLength {
		err = perrors.ErrorfPF("bad length: %d exp %d", len(sqliteText), msTimeStringLength)
		return // bad length: t: zero value err: non-nil
	}

	// parse “2019-01-01 21:30:42.000 +00:00”
	if t, err = time.ParseInLocation(msTimeString, sqliteText, time.UTC); err != nil {
		err = perrors.ErrorfPF("time.Parse: %w", err)
		return // bad parse: t: zero-value err: non-nil
	}
	t = t.Local()

	return // good return: t: time in Local, err: nil
}
