package st

import (
	"errors"
	"fmt"
)

type fatalError struct {
	code int
	error
}

func (f fatalError) ExitStatus() int {
	return f.code
}

// ExitStatuser is an interface for errors that carry an exit status code.
type ExitStatuser interface {
	ExitStatus() int
}

// exitStatus is kept as an alias for internal use.
type exitStatus = ExitStatuser

// Fatal returns an error that will cause stave to print out the
// given args and exit with the given exit code.
func Fatal(code int, args ...interface{}) error {
	return fatalError{
		code:  code,
		error: errors.New(fmt.Sprint(args...)),
	}
}

// Fatalf returns an error that will cause stave to print out the
// given message and exit with the given exit code.
func Fatalf(code int, format string, args ...interface{}) error {
	return fatalError{
		code:  code,
		error: fmt.Errorf(format, args...),
	}
}

// ExitStatus queries the error for an exit status.  If the error is nil, it
// returns 0.  If the error does not implement ExitStatus() int, it returns 1.
// Otherwise it returns the value from ExitStatus().
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	var exit exitStatus
	if errors.As(err, &exit) {
		return exit.ExitStatus()
	}
	return 1
}
