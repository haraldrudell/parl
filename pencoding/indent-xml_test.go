package pencoding_test

import (
	"testing"

	"github.com/haraldrudell/parl/pencoding"
	"github.com/haraldrudell/parl/perrors"
)

func TestIndentXml(t *testing.T) {
	// t.Error("Logging on")
	const (
		errorNo  = false
		errorYes = true

		davIn = `<?xml version="1.0"?>` +
			`<multistatus xmlns="DAV:" xmlns:C="http://x">` +
			`<response>` +
			`<href/>` +
			`<C:x/>` +
			`</response>` +
			`</multistatus>`
		davOut = `<?xml version="1.0"?>
<multistatus xmlns="DAV:" xmlns:C="http://x">
  <response>
    <href/>
    <C:x/>
  </response>
</multistatus>`
	)
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		xmlDocument []byte
		want        []byte
		wantErr     bool
	}{
		{"one element",
			[]byte(`<?xml version="1.0"?><a><b/></a>`),
			[]byte(`<?xml version="1.0"?>` + "\n<a>\n\x20\x20<b/>\n</a>"),
			errorNo},
		{"‘DAV:’", []byte(davIn), []byte(davOut), errorNo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := pencoding.IndentXml(tt.xmlDocument)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("IndentXml() failed: %s", perrors.Short(gotErr))
				}
				return
			}
			if tt.wantErr {
				t.Fatal("IndentXml() succeeded unexpectedly")
			}
			var exp = string(tt.want)
			if s := string(got); s != exp {
				var sS, expS string
				if s == "" {
					sS = "[input: empty]"
				} else {
					sS = s
				}
				if exp == "" {
					expS = "[expected: empty]"
				} else {
					expS = exp
				}
				t.Errorf(
					"FAIL %s: actual %d:\n%s\n%q\n— exp %d:\n%s\n%q\n—",
					tt.name,
					len(s), sS, s,
					len(exp), expS, exp,
				)
			}
		})
	}
}
