/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Port contains an iana port number 1…65535.
package iana

import (
	"github.com/haraldrudell/parl/ints"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

// Port contains an iana port number 1…65535.
//   - Port is ordered
//   - Port implements fmt.Stringer
//   - Port has Uint16, Int and IsValid methods
type Port uint16

func NewPort[T constraints.Integer](integer T) (port Port, err error) {

	// convert to uint16
	var u16 uint16
	if u16, err = ints.ConvertU16(integer, perrors.PackFunc()); err != nil {
		return
	}

	// convert to iana.Port
	port = Port(u16)

	return
}

func NewValidPort[T constraints.Integer](integer T) (port Port, err error) {
	if port, err = NewPort(integer); err != nil {
		return
	}
	if !port.IsValid() {
		err = perrors.ErrorfPF("invalid port value: %d", port)
		return
	}

	return
}

func NewPort1[T constraints.Integer](integer T) (port Port) {
	var err error
	if port, err = NewPort(integer); err != nil {
		panic(err)
	}
	return
}

func (port Port) IsValid() (isValid bool) {
	return port != 0
}

func (port Port) Uint16() (portUint16 uint16) {
	return uint16(port)
}

func (port Port) Int() (portInt int) {
	return int(port)
}
