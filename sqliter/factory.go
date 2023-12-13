/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import "github.com/haraldrudell/parl"

// DSNrFactory is a data-source namer new-function for SQLite3
var DSNrFactory = &dSNrFactory{}

var _ parl.DSNrFactory = &dSNrFactory{}

type dSNrFactory struct{} // empty struct

// NewDSNr returns an object that can
//   - provide data source names from partition selectors and
//   - provide data sources from a data source name
func (df *dSNrFactory) DataSourceNamer(appName string) (dsnr parl.DataSourceNamer, err error) {
	return OpenDataSourceNamer(appName)
}
