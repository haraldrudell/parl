/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"path/filepath"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
)

func TestAppDir(t *testing.T) {
	const appName = "xyzå"
	const localShare = ".local/share"
	// reflects actual state in tester’s homedir
	var pathExp, isExists = func() (pathExp string, isExists bool) {
		var homeDir = UserHomeDir()
		var unevaled = filepath.Join(homeDir, localShare, appName)
		var err error
		if pathExp, err = pfs.AbsEval(unevaled); err == nil {
			isExists = true
			return // appDir exists and was evaled
		}
		if pathExp, err = pfs.AbsEval(filepath.Join(homeDir, localShare)); err == nil {
			pathExp = filepath.Join(pathExp, appName)
			return // parent dir exists and was evaled
		}
		pathExp = unevaled
		return
	}()
	// t.TempDir
	var dir = t.TempDir()
	// TempDir/.local/share/appName
	var pathHookedExp = func() (pathExp string) {
		var err error
		if pathExp, err = pfs.AbsEval(dir); err != nil {
			panic(err)
		}
		pathExp = filepath.Join(pathExp, localShare, appName)
		return
	}()
	var h = homeDirHook
	defer func() { homeDirHook = h }()

	var path string
	var err error
	var isNotExist bool

	// Directory() EnsureDir() Path()
	var appd *AppDirectory = NewAppDir(appName)

	// AppDirectory.App: "xyzå"
	t.Logf("AppDirectory.App: %q", appd.App)

	// App field should be appName
	if appd.App != appName {
		t.Errorf("bad abs %q exp %q", appd.App, appName)
	}

	// Path should return directory path and proper isNotExist
	appd = NewAppDir(appName)
	path, isNotExist, err = appd.Path()

	// real apps
	// path: "/Users/foxyboy/.local/share/xyzå"
	// isNotExist: true
	// err: pfs.AbsEval EvalSymlinks lstat /Users/foxyboy/.local/share/xyzå: no such file or directory at pfs.AbsEval()-abs-eval.go:53
	t.Logf("real apps path: %q isNotExist: %t err: %s",
		path, isNotExist, perrors.Short(err),
	)

	// real errors
	if err != nil && !isNotExist {
		// other error than IsNotExist
		t.Errorf("Path err %s", perrors.Short(err))
	}
	// check isNotExist against isExists
	if isNotExist {
		if isExists {
			t.Error("isNotExist true while isExists true")
		}
	} else if !isExists {
		// directory does not exist
		//	- but isNotExist is not true
		if err == nil {
			t.Error("isNotExist false while isExists false")
		}
	}
	// check path
	if path != pathExp {
		t.Errorf("Path %q exp %q", path, pathExp)
	}

	// hooked Path should return isNotExist true
	appd = NewAppDir(appName)
	homeDirHook = dir
	path, isNotExist, err = appd.Path()
	_ = path
	_ = err
	if !isNotExist {
		t.Error("isNotExist false")
	}

	// hooked EnsureDir should create directory with no error
	appd = NewAppDir(appName)
	homeDirHook = dir
	err = appd.EnsureDir()
	if err != nil {
		t.Errorf("EnsureDir err %s", perrors.Short(err))
	}

	// hooked Path should return abs evaled and isNotExist false
	path, isNotExist, err = appd.Path()
	if err != nil {
		t.Errorf("PathEval err %s", perrors.Short(err))
	}
	if isNotExist {
		t.Error("isNotExist true")
	}
	if path != pathHookedExp {
		t.Errorf("PathEval %q exp %q", path, pathExp)
	}
}

func TestAppDirBadAppName(t *testing.T) {
	var appNameBad = "!"
	var appNameGood = "a"

	var appd *AppDirectory
	var err error
	var path string
	var isNotExist bool

	// good appName no error
	appd = NewAppDir(appNameGood)
	path, isNotExist, err = appd.Path()

	// INFO path "/Users/foxyboy/.local/share/a"
	t.Logf("INFO path %q", path)

	if err != nil && !isNotExist {
		t.Errorf("Path good err %s", perrors.Short(err))
	}

	// bad appname error
	appd = NewAppDir(appNameBad)
	path, isNotExist, err = appd.Path()
	_ = path

	// INFO
	// pos.checkAppName appName can only contain Unicode letters or digits:
	// #0: '!' at pos.(*AppDirectory).checkAppName()-appdir.go:112
	t.Logf("INFO %s", perrors.Short(err))

	if err == nil {
		t.Error("Path missing error")
	}
	if isNotExist {
		t.Error("isNotExist true")
	}
}
