/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io/fs"
	"os"
	"os/exec"
	"path"
	"testing"
)

var noParl = `
package main
func main() {}
`

var withParl = `
package main
import "github.com/haraldrudell/parl"
func main() { _ = parl.ErrEndCallbacks }
`

var withBtree = `
package main
import "github.com/google/btree"
func main() { _ = btree.Ordered }
`

// 67% of parl parse-time is caused by btree
//
// 231125 c66
// Running tool: /opt/homebrew/bin/go test -benchmem -run=^$ -bench ^BenchmarkParl$ github.com/haraldrudell/parl
// goos: darwin
// goarch: arm64
// pkg: github.com/haraldrudell/parl
// BenchmarkParl/no-parl.go-10         	    2454	    484985 ns/op	    5200 B/op	      24 allocs/op
// BenchmarkParl/parl.go-10            	    2486	    515141 ns/op	     30156 parl-ns/op	    5202 B/op	      24 allocs/op
// BenchmarkParl/btree.go-10           	    2467	    505130 ns/op	     20145 btree-ns/op	    5203 B/op	      24 allocs/op
// PASS
// ok  	github.com/haraldrudell/parl	4.940s
func BenchmarkParl(b *testing.B) {
	var benchs = []struct{ filename, metric, code string }{
		{"no-parl.go", "", noParl},
		{"parl.go", "parl-ns/op", withParl},
		{"btree.go", "btree-ns/op", withBtree},
	}
	var uwx fs.FileMode = 0700
	var d = b.TempDir()
	var noParlMetric float64
	for _, bench := range benchs {
		var execCmd = exec.Cmd{Path: path.Join(d, bench.filename)}
		os.WriteFile(execCmd.Path, []byte(bench.code), uwx)
		if bench.metric != "" {
			var preload = execCmd
			preload.Run() // pre-run to ensure parl is stored in ~/go
		}
		b.Run(bench.filename, func(b *testing.B) {
			var x exec.Cmd
			for i := 0; i < b.N; i++ {
				x = execCmd
				x.Run()
			}
			b.StopTimer()
			var elapsed = float64(b.Elapsed()) / float64(b.N)
			if nUnit := bench.metric; nUnit == "" {
				noParlMetric = elapsed
			} else {
				b.ReportMetric(elapsed-noParlMetric, nUnit)
			}
		})
	}
}
