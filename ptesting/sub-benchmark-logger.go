/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptesting

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/ptime"
)

const (
	OperationsPerSecond = "op/s"
)

type SubBenchLogger struct {
	b            *testing.B
	m            map[string]*SubBench
	LastSubBench *SubBench
	T0           time.Time
}

// NewSubBenchLogger returns an object to measure latency of each sub-benchmark b.N invocation
//
// Usage:
//
//	func BenchmarkXxx(b *testing.B) {
//	  benchLogger := NewSubBenchLogger(b)
//	  defer benchLogger.Log()
//	  b.Run(…, func(b *testing.B) {
//	    defer benchLogger.Invo(b)
func NewSubBenchLogger(b *testing.B) (sbl *SubBenchLogger) {
	return &SubBenchLogger{
		b: b,
		m: make(map[string]*SubBench),
	}
}

func (s *SubBenchLogger) Invo(b *testing.B) {
	var subBench *SubBench
	if subBench = s.m[b.Name()]; subBench == nil {
		subBench = &SubBench{Name: b.Name()}
		s.m[subBench.Name] = subBench

		t1 := time.Now()
		if s.LastSubBench != nil {
			s.LastSubBench.D = t1.Sub(s.T0)
		}
		s.LastSubBench = subBench
		s.T0 = t1
	}
	subBench.Runs = append(subBench.Runs, SubBenchRun{BN: b.N, Latency: b.Elapsed()})
}

func (s *SubBenchLogger) Log() {
	if len(s.m) == 0 {
		fmt.Fprintf(os.Stderr, "%s NO sub-benchmark INVOCATIONS\n", s.b.Name())
		return
	}
	if s.LastSubBench != nil {
		t1 := time.Now()
		s.LastSubBench.D = t1.Sub(s.T0)
	}
	for _, subBench := range s.m {
		var maxBN, secondBN string
		length := len(subBench.Runs)
		if length > 0 {
			run := subBench.Runs[length-1]
			maxBN = parl.Sprintf(" max b.N: %s %s",
				BNString(run.BN), ptime.Duration(run.Latency),
			)
		}
		if length > 1 {
			run := subBench.Runs[length-2]
			secondBN = parl.Sprintf(" second b.N: %s %s",
				BNString(run.BN), ptime.Duration(run.Latency),
			)
		}
		bNs := make([]string, length)
		for i, run := range subBench.Runs {
			bNs[i] = BNString(run.BN)
		}
		fmt.Fprintf(os.Stderr, "%s %s %s%s all b.Ns: %s\n",
			subBench.Name, ptime.Duration(subBench.D),
			maxBN, secondBN, strings.Join(bNs, "\x20"),
		)
	}
}

type SubBench struct {
	Name string
	D    time.Duration
	Runs []SubBenchRun
}

type SubBenchRun struct {
	BN      int
	Latency time.Duration
}

// BNString makes a b.N number readable
//   - 1000000 → 1e6
//   - 121293782 → 121,293,782
func BNString(bN int) (s string) {
	s = strconv.Itoa(bN)

	// check if form 1e9 should be used
	if !strings.HasSuffix(s, "00") {
		s = parl.Sprintf("%d", bN) // get string with comma-separators
		return
	}

	exponent := 2
	index := len(s) - exponent - 1
	for index > 0 {
		if s[index] != '0' {
			break
		}
		exponent++
		index--
	}

	s = s[:index+1] + "e" + strconv.Itoa(exponent)
	return
}
