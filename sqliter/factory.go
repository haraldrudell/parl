/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import "github.com/haraldrudell/parl"

// DSNrFactory provides an abstract factory method for an
// SQLite3 data-source namer
var DSNrFactory = &dSNrFactory{}

var _ parl.DSNrFactory = &dSNrFactory{}

type dSNrFactory struct{} // empty struct

// NewDSNr returns an object that can
//   - provide data source names from partition selectors and
//   - provide data sources from a data source name
func (d *dSNrFactory) DataSourceNamer(appName string) (dsnr parl.DataSourceNamer, err error) {
	return OpenDataSourceNamer(appName)
}
