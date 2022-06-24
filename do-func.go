/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// DoThread is invoked in a go statement and executes op.
// g0 receives errors and is the wait-for function.
func DoThread(op func() (err error), g0 Go) {
	var err error
	defer g0.Done(&err)
	defer Recover(Annotation(), &err, NoOnError)

	err = op()
}

func DoProcThread(op func(), g0 Go) {
	var err error
	defer g0.Done(&err)
	defer Recover(Annotation(), &err, NoOnError)

	op()
}

// DoThreadError is a goroutine that returns its error separately.
func DoThreadError(op func() (err error), errCh chan<- error, g0 Go) {
	var err error
	defer g0.Done(&err)
	defer func() {
		errCh <- err
		err = nil
	}()
	defer Recover(Annotation(), &err, NoOnError)

	err = op()
}

// DoGoGetError executes op in a thread.
// err contains any error, error are not submitted to Go object.
// DoGoGetError blocks until the goroutine completes.
func DoGoGetError(op func() (err error), g0 Go) (err error) {
	errCh := make(chan error)
	g0.Add(1)
	go DoThreadError(op, errCh, g0)
	err = <-errCh // block until goroutine completes
	return
}
