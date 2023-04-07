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
//   - Port has IsValid, Int and Uint16 methods
type Port uint16

// NewPort returns iana.Port for any integer value.
//   - values larger that 65535 produce error testable with errors.Is(err, ints.ErrTooLarge)
//   - port may be invalid, ie. not an iana-assigned value, check with port.IsValid
//   - or use NewValidPort
func NewPort[T constraints.Integer](integer T) (port Port, err error) {

	// convert to uint16
	var u16 uint16
	if u16, err = ints.Unsigned[uint16](integer, perrors.PackFunc()); err != nil {
		return
	}

	// convert to iana.Port
	port = Port(u16)

	return
}

// NewValidPort returns iana.Port for any integer value.
//   - values larger that 65535 produce error testable with errors.Is(err, ints.ErrTooLarge)
//   - port is valid
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

// NewPort1 returns iana.Port for any integer value.
//   - if value is too large, panic
//   - port may be invalid, ie. not an iana-allowed value, check with port.IsValid
//   - or use NewValidPort
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
