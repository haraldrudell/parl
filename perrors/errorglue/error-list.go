/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"

	"golang.org/x/exp/slices"
)

// ErrorList returns the list of associated errors enclosed in all error chains of err
//   - the returned slice is a list of error chains beginning with the initial err
//   - If err is nil, a nil slice is returned
//   - duplicate error values are ignored
//   - order is:
//   - — associated errors from err’s error chain, oldest first
//   - — then associated errors from other error chains, oldest found error of each chain first,
//     and errors from the last found error chain first
//   - —
//   - associated errors are independent error chains that allows a single error
//     to enclose addditional errors not part of its error chain
//   - cyclic error values are filtered out so that no error instance is returned more than once
//   - —
//
// ErrorList returns all error-chains reachable from err
//   - err: an error chain to examine, may be nil
//   - errs: a list of error-chains that i each one traversed by itself
//     will reach all erors reachable from err.
//     if err is nil, errs is nil.
//     otherwise, errs[0] is err
//   - —
//   - each traversed error may produce one associated or one or more joined errors
//   - upon obtaining addditional error-chains, traversal of the first error-chain continues while
//     remembering a collection of error-chains for subsequent traversal
//   - this requires either recursion or allocation
//   - [perrors.Errorf] creates [RelatedError], one additional error chain
//   - [errors.Join] creates [JoinUnwrapper], zero or more additional error chains
//   - [fmt.Errorf] creates an error-chain but no additional error-chains
func ErrorList(err error) (errs []error) {

	// nil case
	if err == nil {
		// nil error return: nil slice
		return
	}

	// errMap ensures that no error instance is processed more than once
	var errMap = makeErrIndex()
	// iterator traverses a changing error-chain collection
	//	- the first error-chain traversed is err
	//	- additional error chains may be found during traversal
	var iterator = makeSliceIterator(err)
	// the index in errs for the start of the first error chain
	var chainIndex = 1

	// traverse all error chains reachable from err
	for errorChain := range iterator.iterate {

		// the error at head of error-chain may have been traversed
		// while stored in iterator
		//	- true if errorChain was already traversed
		if errMap.wasTraversed(errorChain) {
			// ignore the error that was already traversed
			continue
		}
		errs = append(errs, errorChain)

		// traverse the current errorChain
		for newErr := errorChain; newErr != nil; {

			// ensure a new error
			if errMap.markTraversed(newErr) {
				// a known error was encountered in the chain
				//	- stop traversing the chain
				break
			}

			// traverse the error chain to the next node
			var associatedError error
			var joinedErrors []error
			if newErr, joinedErrors, associatedError = Unwrap(newErr); associatedError != nil {
				// an associated error from [perrors.AppendError] was found

				// store any associated error not found before
				if errMap.markChain(associatedError) {
					// store the error chain for scanning, first found first
					iterator.addChain(associatedError)
				}
			}

			// process any error-chains from [errors.Join]
			for _, joinedError := range joinedErrors {
				// true: joinedError was not previously encountered
				// as chain or traversed
				if errMap.markChain(joinedError) {
					// store the error chain for traversal in order of occurrence
					iterator.addChain(joinedError)
				}
			}
		}
		// an error chain was iterated to end

		// the chain is ordered newest first
		//	- should be oldest first
		slices.Reverse(errs[chainIndex:])
		chainIndex = len(errs)
	}
	// all error chains were traversed

	return
}

// sliceIterator can iterate a slice that:
//   - is being appended to during iteration
//   - may have its current element removed
type sliceIterator struct {
	// chains is slice that may be modified during iteration
	chains []error
	// i is a field so that methods can modify it
	i int
}

func makeSliceIterator(err error) (s sliceIterator) { return sliceIterator{chains: []error{err}} }

// iterate iterates over a growing slice
func (s *sliceIterator) iterate(yield func(err error) (keepGoing bool)) {
	for s.i < len(s.chains) {
		if !yield(s.chains[s.i]) {
			return
		}
		s.i++
	}
}

func (s *sliceIterator) addChain(errorChain error) {
	fmt.Println("addChain")

	// if sufficient capacity, append at end
	if len(s.chains)+1 <= cap(s.chains) {
		s.chains = s.chains[:len(s.chains)+1]
		s.chains[len(s.chains)-1] = errorChain
		return
	}

	// 0…s.i can be discarded and i set to -1
	//	- +1: then one new element is to be stored
	var newLength = len(s.chains) - (s.i + 1) + 1

	// if at all fits within capacity, copy
	//	- 0…i was already traversed
	if newLength <= cap(s.chains) {
		copy(s.chains, s.chains[s.i+1:])
		s.chains = s.chains[:newLength]
		s.chains[len(s.chains)-1] = errorChain
		s.i = -1
		return
	}

	// allocation: extend capacity using aappend
	s.chains = append(s.chains, errorChain)
}

// errIndex kees track of traversed errors and
// errors identified as additional error chains
type errIndex map[error]mapStatus

// makeErrIndex returns an index of identified error values
func makeErrIndex() (e errIndex) { return make(map[error]mapStatus) }

// markChain checks whether err is already known
//   - isNew true: err was not previously known.
//     err was marked as chain
func (i errIndex) markChain(err error) (isNew bool) {
	if _, isKnown := i[err]; isKnown {
		// error exists in index: isNew false
		return
	}
	i[err] = isChain
	isNew = true

	return
}

// wasTraversed returns true if err was already traversed
func (i errIndex) wasTraversed(err error) (was bool) {
	var status, isKnown = i[err]
	was = isKnown && status == wasTraversed

	return
}

// markTraversed marks err as traversed
//   - was: true if err was already traversed
func (i errIndex) markTraversed(err error) (was bool) {
	var status, isKnown = i[err]
	if isKnown && status == wasTraversed {
		was = true
		return
	}
	i[err] = wasTraversed

	return
}

const (
	// the error was stored as an error chain
	isChain mapStatus = iota + 1
	// the error was traversed
	wasTraversed
)

// [isChain] [wasTraversed]
type mapStatus uint8
