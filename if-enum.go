/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/iters"
	"golang.org/x/exp/constraints"
)

// Enum is an enumeration using string keys mapping to a value type T.
// T is a unique named type that separates the values of this enumeration
// from all other values.
type Enum[T any] interface {
	KeyEnum[string, T]
}

// KeyedEnum is an enumeration using a key type K mapping to a value type T.
//
//   - T is a unique named type that separates the values of this enumeration from all other values.
//   - K is a common type, often a single-word string, whose values are unique and maps one-to-one
//     to a T value. K is constraints.Ordered and can be used as a map key.
//   - The implementation’s stored type may be different from both K and T.
//
// Some benefits with enumerations are:
//   - unknown, illegal or value duplications are detected
//   - integral values and their meanings can be printed
//   - one set of integral values are not confused with another, eg. unix.AF_INET and
//     unix.RTAX_BRD
//   - Reveals the meaning when used as function arguments and other allowed
//     values can be examined
type KeyEnum[K constraints.Ordered, T any] interface {

	// K

	IsKey(key K) (isKey bool)                  // IsKey checks whether key maps to an enumerated value
	Value(key K) (enum T, err error)           // Value looks up an enumerated value by key
	KeyIterator() (iterator iters.Iterator[K]) // KeyIterator returns an iterator that iterates over all keys in order of definition

	// T

	IsValid(enum T) (isEnumValue bool)      // IsValid checks whether value is among enumerated values
	Key(value T) (key K, err error)         // Key gets the key value for an enumerated T value
	ValueAny(value any) (enum T, err error) // ValueAny attempts to convert any value to a T enumerated value
	Iterator() (iterator iters.Iterator[T]) // Iterator returns an iterator that iterates over all enumerated values in order of definition
	Description(enum T) (desc string)       // Description gets a descriptive sentence for an enum value
	StringT(enum T) (s string)              // StringT provides a string representation for an enumeration value
	// Compare compares two T values.
	//	- result is 0 if the two values are considered equal
	//	- result is 1 if value1 is considered greater than value2
	//	- result is -1 if value1 is considered less than value2
	Compare(value1, value2 T) (result int)

	Name() (s string) // Name returns a short string naming this enumeration
	fmt.Stringer
}

// EnumItem is a generic interface for enumeration item implementations.
// Enumeration items are ordered by the K key type.
//   - K is a key type whose values map to restricted type V values one-to-one.
//   - V is a restricted type for enumeration values that may store more efficiently compared to a portable type.
type EnumItem[K constraints.Ordered, V any] interface {
	Key() (key K)               // Key returns the key for this enumeration value
	Description() (desc string) // Description returns a descriptive sentence for this enumeration value
	Value() (value V)           // Value returns this enumeration value’s value using the restricted type

	fmt.Stringer
}
