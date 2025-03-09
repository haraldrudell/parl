/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/sets"
)

// GoFlow enforces new, pre-go and go in sequence
type GoFlow struct {
	// state is state of GoFlow
	state Atomic32[GoFlowState]
}

// NewGoFlow initializes the GoFlow pointed to by fieldp
//   - toState missing: sets GoFlow to initialized state
//   - toState [GPrego]: makes GoFlow ready for go statement
//   - —
//   - with PrepareToGo: [NewGoFlow] [GoFlow.PrepareToGo] [GoFlow.Go]
//   - only Go: [NewGoFlow](GPrego) [GoFlow.Go]
//
// Usage:
//
//	type S struct {
//	  goFlow GoFlow
//	  wg sync.WaitGroup
//	  …
//	func NewS(…) (…) {
//	  NewGoFlow(&s.goFlow)
//	  …
//	func (s *S) PreGo(…) {
//	  s.goFlow.Prego()
//	  s.wg.Add(1)
//	  …
//	func (s * S) Goroutine(…) {
//	  defer s.wg.Done()
//	  s.goFlow.Go()
func NewGoFlow(toState ...GoFlowState) (goFlow *GoFlow) {
	return (&GoFlow{}).Initialize(toState...)
}

// Initialize takes existing GoFlow from uninitialized to initialized
//   - toState missing: sets GoFlow to initialized state
//   - toState [GPrego]: makes GoFlow ready for go statement
//   - —
//   - with PrepareToGo: [GoFlow.Initialize] [GoFlow.PrepareToGo] [GoFlow.Go]
//   - only Go: [GoFlow.Initialize](GPrego) [GoFlow.Go]
//   - thread-safe, supports functional-chaining
func (g *GoFlow) Initialize(toState ...GoFlowState) (gf *GoFlow) {

	// get nextState and check toState
	var nextState GoFlowState
	if len(toState) == 0 {
		nextState = gInitialized
	} else {
		nextState = GPrego
		if s := toState[0]; s != GPrego {
			panic(perrors.ErrorfPF("initialize toState bad %s expected %s",
				s, GPrego,
			))
		}
	}

	// verify GoFlow state
	switch state := g.state.Load(); state {
	case GUninitialized:
	case gInitialized:
		panic(perrors.NewPF("multiple Initialize or new-function"))
	default:
		panic(perrors.ErrorfPF("bad state for initialize %s exp %s",
			state, GUninitialized,
		))
	}

	// transition to nextState
	if g.state.CompareAndSwap(GUninitialized, nextState) {
		gf = g
		return
	}

	var state = g.state.Load()
	panic(perrors.ErrorfPF("race condition at initialize state now %s was %s",
		state, GUninitialized,
	))
}

// PrepareToGo is invoked before go statement
//   - GoFlow must be initialized
//   - PrepareToGo can only be invoked once
//   - next action is [GoFlow.Go]
func (g *GoFlow) PrepareToGo() {

	// verify GoFlow state
	switch state := g.state.Load(); state {
	case gInitialized:
	case GUninitialized:
		panic(perrors.NewPF("uninitialized GoFlow"))
	case GPrego:
		panic(perrors.NewPF("PrepareToGo more than once"))
	default:
		panic(perrors.ErrorfPF("corrupt state %s exp %s", state, gInitialized))
	}

	// transition to GPrego
	if g.state.CompareAndSwap(gInitialized, GPrego) {
		return // success
	}

	var state = g.state.Load()
	panic(perrors.ErrorfPF("race condition at PrepareToGo state now %s was %s",
		state, gInitialized,
	))
}

// Go is invoked in the goroutine function
//   - possible paths:
//   - [GoFlow.Go](GUninitialized)
//   - [NewGoFlow] [GoFlow.PrepareToGo] [GoFlow.Go]
//   - [NewGoFlow](GPrego) [GoFlow.Go]
//   - [GoFlow.Initialize] [GoFlow.PrepareToGo] [GoFlow.Go]
//   - [GoFlow.Initialize](GPrego) [GoFlow.Go]
//   - Go can only be invoked once
func (g *GoFlow) Go(isState ...GoFlowState) {

	// get expected state
	var shouldBeState GoFlowState
	if len(isState) == 0 {
		shouldBeState = GPrego
	} else {
		shouldBeState = GUninitialized
		if s := isState[0]; s != GUninitialized {
			panic(perrors.ErrorfPF("Go isState bad %s expected %s",
				s, GUninitialized,
			))
		}
	}

	// verify GoFlow state
	switch state := g.state.Load(); state {
	case GUninitialized:
		if state == shouldBeState {
			break
		}
		panic(perrors.NewPF("uninitialized GoFlow in Go"))
	case GPrego:
		if state == shouldBeState {
			break
		}
		panic(perrors.ErrorfPF("Go GloFlow state bad %s expected %s",
			state, shouldBeState,
		))
	case shouldBeState:
	case gInitialized:
		panic(perrors.NewPF("Go without PrepareToGo"))
	case ggo:
		panic(perrors.NewPF("Go more than once"))
	default:
		panic(perrors.ErrorfPF("corrupt state in Go: %s exp %s",
			state, shouldBeState,
		))
	}

	// transition state
	if g.state.CompareAndSwap(shouldBeState, ggo) {
		return // success
	}

	var state = g.state.Load()
	panic(perrors.ErrorfPF("race condition in Go state-change from %s to %s",
		shouldBeState, state,
	))
}

const (
	// the GoFlow has not been subject to new-function
	GUninitialized GoFlowState = iota
	// the GoFlow was initialized by new-function
	gInitialized
	// the GoFlow did new-function and pre-go
	GPrego
	// the goroutine is running
	ggo
)

// GoFlow state: gInitialized [GPrego] ggo gUninitialized
type GoFlowState uint32

func (g GoFlowState) String() (s string) { return goFlowSet.StringT(g) }

// goFlowSet is string translation for GoFlowState
var goFlowSet = sets.NewSet[GoFlowState]([]sets.SetElement[GoFlowState]{
	{ValueV: GUninitialized, Name: "uninitialized"},
	{ValueV: gInitialized, Name: "initialized"},
	{ValueV: GPrego, Name: "ready-to-go"},
	{ValueV: ggo, Name: "goroutine-running"},
})
