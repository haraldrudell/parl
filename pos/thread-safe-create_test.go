/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateFile(t *testing.T) {
	var (
		tempDir  = t.TempDir()
		filename = filepath.Join(tempDir, "x.txt")
	)

	var (
		osFile  *os.File
		isExist bool
		err     error
	)

	// create success
	osFile, isExist, err = CreateFile(filename)
	if err != nil {
		t.Errorf("CreateFile err: %s", err)
	}
	if isExist {
		t.Error("isExist true")
	}
	if osFile == nil {
		t.Fatal("osFile nil")
	}
	err = osFile.Close()
	if err != nil {
		t.Errorf("Close err: %s", err)
	}

	// create when exists
	osFile, isExist, err = CreateFile(filename)
	if err == nil {
		t.Error("missing error")
	} else if !errors.Is(err, os.ErrExist) {
		t.Errorf("CreateFile err: %s", err)
	}
	if !isExist {
		t.Error("isExist false")
	}
	if osFile != nil {
		t.Fatal("osFile non-nil")
	}
}

// ITEST= go test -run ^TestCreateFile_Integration$ github.com/haraldrudell/parl/pos
func TestCreateFile_Integration(t *testing.T) {

	// check environment
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("ITEST not present")
	}

	const (
		f = "/opt/f/meta/2025/x"
	)
	var (
		osFile  *os.File
		isExist bool
		err     error
		n       int
		// 2006-01-02T15:04:05Z07:00
		text      = time.Now().Format(time.RFC3339)
		expLength = len(text)
		contents  []byte
	)

	// create should succeed
	t.Logf("creating %q…", f)
	osFile, isExist, err = CreateFile(f)
	if isExist {
		t.Errorf("File exists: %q", f)
	}
	if err != nil {
		t.Fatal(err)
	}

	// write should be 4 bytes
	n, err = osFile.WriteString(text)
	if err != nil {
		t.Errorf("WriteString err: %s", err)
	}
	if n != expLength {
		t.Errorf("Bad write: %d exp %d", n, expLength)
	}

	// close should succeed
	err = osFile.Close()
	if err != nil {
		t.Errorf("Close err: %s", err)
	}

	// verify content
	contents, err = os.ReadFile(f)
	if err != nil {
		t.Errorf("ReadFile err: %s", err)
	}
	if s := string(contents); s != text {
		t.Errorf("bad contents:\n%q exp\n%q", s, text)
	}
}
