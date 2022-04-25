/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import "github.com/haraldrudell/parl"

var DSNrFactory = &dSNrFactory{}

type dSNrFactory struct{}

func (df *dSNrFactory) NewDSN(appName string) (dsnr parl.DataSourceNamer) {
	return NewDataSourceNamer(appName)
}
