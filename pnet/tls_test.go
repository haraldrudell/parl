/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pnet

/*
func TestTls(t *testing.T) {
	/*
	TODO delete file
			to set test verbosity in Visual Studio Code 1.59.0 210814 on macOS:
			macOS menu bar — Code — Preferences — Settings (or command+,)
			under “Go: Test Flags”, click “Edit in settings.json”
			edit the key to be: "go.testFlags": ["-v"]
		— test results are cached, so a change is required to actually run the test
		— the -v flag makes t.Log* to be printed
		— printouts are streamed, first appears after about 1s
	socketName := "127.0.0.1:0"
	hp := NewHttp(socketName)
	errCh := hp.Run()
	var wg sync.WaitGroup
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		var ok bool
		err, ok = <-errCh
		_ = ok
	}()
	t.Logf("Address: %s", hp.Addr.String())
	wg.Wait()
	if err != nil {
		t.Errorf("http error: %+v", err)
	}
}
*/
