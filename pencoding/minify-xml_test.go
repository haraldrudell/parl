package pencoding_test

import (
	"testing"

	"github.com/haraldrudell/parl/pencoding"
)

func TestMinifyXml(t *testing.T) {
	// t.Error("Logging on")
	const (
		errorNo  = false
		errorYes = true

		davIn = `<?xml version="1.0"?>
<multistatus xmlns="DAV:" xmlns:C="http://x">
  <response>
    <href/>
    <C:x/>
  </response>
</multistatus>`
		davOut = `<?xml version="1.0"?>` +
			`<multistatus xmlns="DAV:" xmlns:C="http://x">` +
			`<response>` +
			`<href/>` +
			`<C:x/>` +
			`</response>` +
			`</multistatus>`
	)
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		xmlDoc  []byte
		want    []byte
		wantErr bool
	}{
		{"nil", nil, nil, errorNo},
		{"space", []byte("\x20"), nil, errorNo},
		{"one element", []byte("<x></x>"), []byte("<x/>"), errorNo},
		{"self-closing element", []byte("<x/>"), []byte("<x/>"), errorNo},
		{"indented", []byte("<x>\x20\x20<y/></x>"), []byte("<x><y/></x>"), errorNo},
		{"double-quote", []byte(`<x>"</x>`), []byte(`<x>"</x>`), errorNo},
		// less-than and ampersand must be quoted
		// encoding/xml also quotes double-quote greater-than
		{"difficult chars",
			[]byte(`<x>"&amp;&lt;'>` + "\ta\x0aad\x0dd" + `</x>`),
			[]byte("<x>&#34;&amp;&lt;&#39;&gt;&#x9;a\x0aad\x0ad</x>"), errorNo},
		{"‘DAV:’", []byte(davIn), []byte(davOut), errorNo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := pencoding.MinifyXml(tt.xmlDoc)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("TrimXml() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("TrimXml() succeeded unexpectedly")
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
