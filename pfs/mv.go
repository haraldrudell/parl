/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"os/exec"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	mvCommand   = "mv"
	mvNoClobber = "-n"
)

// Mv moves a file or directory structure using system command.
// Mv uses -n for --n-clobber.
// Mv does not indicate if moive was aborted due to no-clobber.
// outConsumer can be nil but receiver command output if any.
// mv typically has no output.
func Mv(src, dest string, outConsumer func(string)) (err error) {
	parl.Debug("Mv src: %s dest: %s\n", src, dest)
	var bytes []byte
	bytes, err = exec.Command(mvCommand, mvNoClobber, src, dest).CombinedOutput()
	if len(bytes) != 0 && outConsumer != nil {
		outConsumer(string(bytes))
	}
	if err != nil {
		err = perrors.Errorf("exec.Command mv: '%w'", err)
	}
	return
}
