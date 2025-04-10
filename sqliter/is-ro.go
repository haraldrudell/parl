/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"github.com/haraldrudell/parl"
)

// IsRO returns read-only flag for any data source namer
func IsRO(dataSourceNamer parl.DataSourceNamer) (isRO parl.ROtype) {

	// check for type with indicator
	if dsnr, isROType := dataSourceNamer.(parl.IsRoDsnr); isROType {
		isRO = dsnr.IsRO()
		return
	}

	// check for RO type
	if _, isROType := dataSourceNamer.(*DataSourceNamerRO); isROType {
		isRO = parl.ROyes
		return
	}

	// must be non-RO type
	//	- includes [*DataSourceNamer]
	isRO = parl.ROno

	return
}
